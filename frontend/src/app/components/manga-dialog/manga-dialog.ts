import { Component, model, output, inject, signal, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { DialogModule } from 'primeng/dialog';
import { MangaSearchSelectComponent, MangaSearchResult } from '../search-select/search-select';
import { CronInputComponent } from '../cron-input/cron-input';
import { environment } from 'app/environments/environment';

const ANILIST_QUERY = `
  query ($search: String) {
    Media(type: MANGA, search: $search) {
      coverImage {
        extraLarge
        large
        medium
      }
    }
  }
`;

@Component({
  selector: 'app-add-manga-dialog',
  standalone: true,
  imports: [CommonModule, FormsModule, DialogModule, MangaSearchSelectComponent, CronInputComponent],
  templateUrl: './manga-dialog.html',
  styleUrls: ['./manga-dialog.css'],
})
export class AddMangaDialogComponent {
  @ViewChild(MangaSearchSelectComponent) private mangaSelect!: MangaSearchSelectComponent;

  visible = model<boolean>(false);
  mangaAdded = output<void>();

  private http = inject(HttpClient);

  selectedManga: MangaSearchResult | null = null;
  coverUrl = signal<string | null>(null);
  coverLoading = signal(false);
  cron = '0 * * * *';
  mangaPath = '';
  submitting = signal(false);
  error = signal<string | null>(null);

  onMangaSelected(manga: MangaSearchResult): void {
    this.selectedManga = manga;
    this.mangaPath = `./mangas/${manga.name.toLowerCase().replace(/[^a-z0-9]+/g, '_')}`;
    this.fetchAnilistCover(manga.name);
  }

  private fetchAnilistCover(name: string): void {
    this.coverLoading.set(true);
    this.coverUrl.set(null);

    this.http.post<any>('https://graphql.anilist.co', {
      query: ANILIST_QUERY,
      variables: { search: name },
    }, {
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    }).subscribe({
      next: (res) => {
        const cover = res?.data?.Media?.coverImage;
        this.coverUrl.set(cover?.extraLarge ?? cover?.large ?? cover?.medium ?? null);
        this.coverLoading.set(false);
      },
      error: () => {
        this.coverUrl.set(null);
        this.coverLoading.set(false);
      },
    });
  }

  onMangaCleared(): void {
    this.selectedManga = null;
    this.mangaPath = '';
    this.coverUrl.set(null);
  }

  onConfirm(): void {
    if (!this.selectedManga) return;

    this.submitting.set(true);
    this.error.set(null);

    const body = {
      name: this.selectedManga.name,
      provider: this.selectedManga.provider,
      time_rule: this.cron,
      url: this.selectedManga.url,
      manga_path: this.mangaPath,
    };

    this.http.post<{ manga_id: number }>(`${environment.backendUrl}/add/manga`, body).subscribe({
      next: (res) => {
        this.submitting.set(false);
        if (res?.manga_id) {
          this.mangaAdded.emit();
          this.close();
        } else {
          this.error.set('Unexpected response from server.');
        }
      },
      error: (err) => {
        this.submitting.set(false);
        this.error.set(err?.error?.message ?? 'Something went wrong. Please try again.');
      },
    });
  }

  reset(): void {
    this.selectedManga = null;
    this.cron = '0 * * * *';
    this.mangaPath = '';
    this.error.set(null);
    this.submitting.set(false);
    this.coverUrl.set(null);
    this.coverLoading.set(false);
    this.mangaSelect?.clear();
  }

  close(): void {
    this.visible.set(false);
  }
}