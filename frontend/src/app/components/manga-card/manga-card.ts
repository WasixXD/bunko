import { Component, Input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TagModule } from 'primeng/tag';
import { CardModule } from 'primeng/card';
import { Manga } from '../../mangas/mangas.model' 

@Component({
  selector: 'app-manga-card',
  standalone: true,
  imports: [CommonModule, TagModule, CardModule],
  templateUrl: './manga-card.html',
  styleUrl: './manga-card.css'
})
export class MangaCardComponent {
  @Input() manga!: Manga;

  clicked = output<number>();

  get statusSeverity(): 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast' {
    const map: Record<string, 'success' | 'info' | 'warn' | 'secondary'> = {
      completed: 'success',
      downloading: 'info',
      pending: 'warn',
    };
    return map[this.manga.status?.toLowerCase()] ?? 'secondary';
  }

  get displayTitle(): string {
    return this.manga.localized_name || this.manga.name;
  }

  get coverUrl(): string {
    return this.manga.cover_path ?? 'assets/no-cover.png';
  }

  onClick(): void {
    this.clicked.emit(this.manga.manga_id);
  }
}
