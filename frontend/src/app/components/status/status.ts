import { Component, OnInit, OnDestroy, signal, computed, input } from '@angular/core';
import { environment } from 'app/environments/environment';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Manga } from 'app/mangas/mangas.model';

type JobStatus = 'downloading' | 'pending' | 'completed' | 'error';

interface Job {
  manga_id: number;
  name: string;
  path_to_download: string;
  provider: string;
  rowid: number;
  status: JobStatus;
  url: string;
}

@Component({
  selector: 'app-status',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './status.html',
  styleUrl: './status.css',
})
export class StatusComponent implements OnInit, OnDestroy {
  mangas = input<Manga[]>([]);

  private source = new EventSource(`${environment.backendUrl}/queue`);
  jobs = signal<Job[]>([]);
  activeFilter = signal<JobStatus | 'all'>('all');
  nameQuery = signal('');

  readonly filters: Array<{ label: string; value: JobStatus | 'all' }> = [
    { label: 'All', value: 'all' },
    { label: 'Downloading', value: 'downloading' },
    { label: 'Pending', value: 'pending' },
    { label: 'Completed', value: 'completed' },
    { label: 'Error', value: 'error' },
  ];

  filteredJobs = computed(() => {
    const filter = this.activeFilter();
    const query = this.nameQuery().trim().toLowerCase();
    return this.jobs().filter((j) => {
      const matchesStatus = filter === 'all' || j.status === filter;
      const matchesName = !query || j.name.toLowerCase().includes(query);
      return matchesStatus && matchesName;
    });
  });

  counts = computed(() => {
    const all = this.jobs() ?? [];
    return {
      all: all.length,
      downloading: all.filter((j) => j.status === 'downloading').length,
      pending: all.filter((j) => j.status === 'pending').length,
      completed: all.filter((j) => j.status === 'completed').length,
      error: all.filter((j) => j.status === 'error').length,
    };
  });

  mangaExists(mangaId: number): boolean {
    return (this.mangas() ?? []).some((m) => m.manga_id === mangaId);
  }

  getCover(mangaId: number): string | null {
    return (this.mangas() ?? []).find((m) => m.manga_id === mangaId)?.cover_path ?? null;
  }

  ngOnInit() {
    this.source.onmessage = (event) => {
      const result = JSON.parse(event.data);
      this.jobs.set(result ?? []);
    };
  }

  ngOnDestroy() {
    this.source.close();
  }

  setFilter(value: JobStatus | 'all') {
    this.activeFilter.set(value);
  }

  onNameInput(value: string) {
    this.nameQuery.set(value);
  }
}