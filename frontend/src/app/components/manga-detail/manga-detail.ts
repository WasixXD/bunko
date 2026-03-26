import { Component, effect, inject, input, model, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DialogModule } from 'primeng/dialog';
import { Manga } from 'app/mangas/mangas.model';
import { MangaService } from 'app/mangas/mangas.service';

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

  private mangaService = inject(MangaService);

  manga = signal<Manga | null>(null);
  loading = signal(false);
  error = signal<string | null>(null);
  confirmingDelete = signal(false);
  deleting = signal(false);
  updatingMetadata = signal(false);
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

    this.mangaService.getMangaById(id).subscribe({
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

  get metadataUpdatedAt(): string {
    const value = this.manga()?.metadata_updated_at;
    if (!value) return 'Never';

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return new Intl.DateTimeFormat(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(date);
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

    this.mangaService.deleteManga(id).subscribe({
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

  updateMetadata(id: number): void {
    this.updatingMetadata.set(true);
    this.error.set(null);

    this.mangaService.updateMetadata(id).subscribe({
      next: (manga) => {
        this.manga.set(manga);
        this.updatingMetadata.set(false);
      },
      error: () => {
        this.error.set('Failed to update metadata.');
        this.updatingMetadata.set(false);
      },
    });
  }
}
