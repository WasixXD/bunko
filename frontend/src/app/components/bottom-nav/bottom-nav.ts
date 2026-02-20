import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AddMangaDialogComponent } from '../manga-dialog/manga-dialog';

export type AppPage = 'home' | 'status';

@Component({
  selector: 'app-bottom-nav',
  standalone: true,
  imports: [CommonModule, AddMangaDialogComponent],
  templateUrl: './bottom-nav.html',
  styleUrls: ['./bottom-nav.css'],
})
export class BottomNavComponent {
  activePage = input.required<AppPage>();
  pageChange = output<AppPage>();
  mangaAdded = output<void>();

  addDialogVisible = signal(false);

  navigate(page: AppPage): void {
    this.pageChange.emit(page);
  }

  openAddDialog(): void {
    this.addDialogVisible.set(true);
  }

  onMangaAdded(): void {
    this.mangaAdded.emit();
  }
}