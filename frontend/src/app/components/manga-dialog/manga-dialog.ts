import { Component, model, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { MangaSearchSelectComponent, MangaSearchResult } from '../search-select/search-select';
import { CronInputComponent } from '../cron-input/cron-input';

@Component({
  selector: 'app-add-manga-dialog',
  standalone: true,
  imports: [CommonModule, FormsModule, DialogModule, MangaSearchSelectComponent, CronInputComponent],
  templateUrl: './manga-dialog.html',
  styleUrls: ['./manga-dialog.css'],
})
export class AddMangaDialogComponent {
  visible = model<boolean>(false);
  confirmed = output<{ manga: MangaSearchResult; cron: string; path: string }>();

  selectedManga: MangaSearchResult | null = null;
  cron = '0 * * * *';
  mangaPath = '';

  onMangaSelected(manga: MangaSearchResult): void {
    this.selectedManga = manga;
    this.mangaPath = `./mangas/${manga.name.toLowerCase().replace(/[^a-z0-9]+/g, '_')}`;
  }

  onMangaCleared(): void {
    this.selectedManga = null;
    this.mangaPath = '';
  }

  onConfirm(): void {
    if (!this.selectedManga) return;
    this.confirmed.emit({ manga: this.selectedManga, cron: this.cron, path: this.mangaPath });
    this.close();
  }

  close(): void {
    this.visible.set(false);
    this.selectedManga = null;
    this.cron = '0 * * * *';
    this.mangaPath = '';
  }
}