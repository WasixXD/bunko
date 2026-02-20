import { CommonModule } from '@angular/common';
import { Component, OnInit, signal, inject, DestroyRef } from '@angular/core';
import { Manga } from 'app/mangas/mangas.model';
import { MangaService } from 'app/mangas/mangas.service';
import { MangaCardComponent } from '@components/manga-card/manga-card';
import { BottomNavComponent, AppPage } from '@components/bottom-nav/bottom-nav';
import { StatusComponent } from '@components/status/status';
import { ToastModule } from 'primeng/toast';
import { MessageService } from 'primeng/api';
import { interval, Subscription, switchMap, takeWhile, tap } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

const POLL_INTERVAL_MS = 3000;
const POLL_MAX_ATTEMPTS = 10;

@Component({
  selector: 'app-home-page',
  imports: [CommonModule, MangaCardComponent, BottomNavComponent, StatusComponent, ToastModule],
  standalone: true,
  providers: [MessageService],
  templateUrl: './home.page.html',
  styleUrl: './home.page.css',
})
export class HomePage implements OnInit {
  mangas = signal<Array<Manga>>([]);
  activePage = signal<AppPage>('home');

  private mangaService = inject(MangaService);
  private messageService = inject(MessageService);
  private destroyRef = inject(DestroyRef);
  private pollSub: Subscription | null = null;
  private pollAttempts = 0;

  ngOnInit() {
    this.loadMangas();
  }

  onPageChange(page: AppPage): void {
    this.activePage.set(page);
    if (page === 'home' && this.mangas().length === 0) {
      this.loadMangas();
    }
  }

  onMangaAdded(): void {
    this.messageService.add({
      severity: 'success',
      summary: 'Success',
      detail: 'Manga added! Loading metadata...',
      life: 4000,
    });

    this.stopPolling();
    this.pollAttempts = 0;

    this.pollSub = interval(POLL_INTERVAL_MS)
      .pipe(
        takeUntilDestroyed(this.destroyRef),
        switchMap(() => this.mangaService.getMangas()),
        tap((data) => this.mangas.set(data)),
        takeWhile((data) => {
          this.pollAttempts++;
          const allHaveCovers = data.every((m) => !!m.cover_path);
          const maxReached = this.pollAttempts >= POLL_MAX_ATTEMPTS;
          if (allHaveCovers || maxReached) {
            if (allHaveCovers) {
              this.messageService.add({
                severity: 'success',
                summary: 'Ready',
                detail: 'Metadata loaded successfully.',
                life: 3000,
              });
            }
            return false;
          }
          return true;
        })
      )
      .subscribe();
  }

  private loadMangas(): void {
    this.mangaService.getMangas().subscribe({
      next: (data) => this.mangas.set(data),
      error: (err) => console.error(err),
    });
  }

  private stopPolling(): void {
    this.pollSub?.unsubscribe();
    this.pollSub = null;
  }
}