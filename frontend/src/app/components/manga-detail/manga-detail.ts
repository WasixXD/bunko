import { Component, model, input, inject, signal, effect, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { DialogModule } from 'primeng/dialog';
import { Manga } from 'app/mangas/mangas.model';
import { environment } from 'app/environments/environment';

@Component({
  selector: 'app-manga-detail-dialog',
  standalone: true,
  imports: [CommonModule, DialogModule],
  templateUrl: './manga-detail.html',
  styleUrl: './manga-detail.css',
})
export class MangaDetailDialogComponent {
  visible = model<boolean>(false);
  mangaId = input<number | null>(null);

  private http = inject(HttpClient);

  manga = signal<Manga | null>(null);
  loading = signal(false);
  error = signal<string | null>(null);
  confirmingDelete = signal(false);
  deleting = signal(false);
  mangaDeleted = output<void>();

  constructor() {
    effect(() => {
      const id = this.mangaId();
      if (id != null && this.visible()) {
        this.fetchManga(id);
      }
    });
  }

  private fetchManga(id: number): void {
    this.loading.set(true);
    this.error.set(null);
    this.manga.set(null);

    this.http.get<Manga>(`${environment.backendUrl}/mangas/get/?id=${id}`).subscribe({
      next: (data) => {
        this.manga.set(data);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Failed to load manga details.');
        this.loading.set(false);
      },
    });
  }

  get cleanSummary(): string {
    return (this.manga()?.summary ?? '')
      .replace(/<br\s*\/?>/gi, '\n')
      .replace(/<[^>]+>/g, '')
      .trim();
  }

  get startDate(): string {
    const m = this.manga();
    if (!m?.start_year) return '—';
    const parts: string[] = [String(m.start_year)];
    if (m.start_month) parts.push(String(m.start_month).padStart(2, '0'));
    if (m.start_day) parts.push(String(m.start_day).padStart(2, '0'));
    return parts.join(' / ');
  }

  close(): void {
    this.visible.set(false);
    this.manga.set(null);
    this.confirmingDelete.set(false);
  }

  requestDelete(): void {
    this.confirmingDelete.set(true);
  }

  cancelDelete(): void {
    this.confirmingDelete.set(false);
  }

  confirmDelete(id: number): void {
    this.deleting.set(true);
    this.error.set(null);

    this.http.delete(`${environment.backendUrl}/mangas/delete/?id=${id}`).subscribe({
      next: () => {
        this.deleting.set(false);
        this.visible.set(false);
        this.mangaDeleted.emit();
      },
      error: () => {
        this.error.set('Failed to delete manga.');
        this.deleting.set(false);
        this.confirmingDelete.set(false);
      },
    });
  }
}