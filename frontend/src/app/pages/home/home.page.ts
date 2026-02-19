import { CommonModule } from '@angular/common';
import { Component, OnInit, signal } from '@angular/core';
import { Manga } from 'app/mangas/mangas.model';
import { MangaService } from 'app/mangas/mangas.service';
import { MangaCardComponent } from '@components/manga-card/manga-card';
import { BottomNavComponent, AppPage } from '@components/bottom-nav/bottom-nav';
import { StatusComponent } from '@components/status/status';

@Component({
  selector: 'app-home-page',
  imports: [CommonModule, MangaCardComponent, BottomNavComponent, StatusComponent],
  standalone: true,
  templateUrl: './home.page.html',
  styleUrl: './home.page.css',
})
export class HomePage implements OnInit {
  mangas = signal<Array<Manga>>([]);
  activePage = signal<AppPage>('home');

  constructor(private mangaService: MangaService) {}

  ngOnInit() {
    this.loadMangas();
  }

  onPageChange(page: AppPage): void {
    this.activePage.set(page);
    if (page === 'home' && this.mangas().length === 0) {
      this.loadMangas();
    }
  }

  private loadMangas(): void {
    this.mangaService.getMangas().subscribe({
      next: (data) => this.mangas.set(data),
      error: (err) => console.error(err),
    });
  }
}