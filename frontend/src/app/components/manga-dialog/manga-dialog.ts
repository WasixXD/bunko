import { Component, model, output, inject, signal, ViewChild, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { Subject, Subscription, finalize, forkJoin, of } from 'rxjs';
import { catchError, debounceTime, distinctUntilChanged, switchMap } from 'rxjs/operators';
import { DialogModule } from 'primeng/dialog';
import { MangaSearchSelectComponent, MangaSearchResult } from '../search-select/search-select';
import { CronInputComponent } from '../cron-input/cron-input';
import { environment } from 'app/environments/environment';
import { MangaService, PathSuggestion, PathSuggestionResponse, PathValidationResponse } from 'app/mangas/mangas.service';

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
export class AddMangaDialogComponent implements OnInit, OnDestroy {
  @ViewChild(MangaSearchSelectComponent) private mangaSelect!: MangaSearchSelectComponent;

  visible = model<boolean>(false);
  mangaAdded = output<void>();

  private http = inject(HttpClient);
  private mangaService = inject(MangaService);

  selectedManga: MangaSearchResult | null = null;
  coverUrl = signal<string | null>(null);
  coverLoading = signal(false);
  cron = '0 * * * *';
  mangaPath = '';
  submitting = signal(false);
  error = signal<string | null>(null);
  pathValidation = signal<PathValidationResponse | null>(null);
  pathChecking = signal(false);
  pathSuggestions = signal<PathSuggestion[]>([]);
  showSuggestions = signal(false);
  activeSuggestionIndex = signal(-1);

  private readonly pathInput$ = new Subject<string>();
  private pathInputSub?: Subscription;

  ngOnInit(): void {
    this.pathInputSub = this.pathInput$
      .pipe(
        debounceTime(250),
        distinctUntilChanged(),
        switchMap((path) => {
          const trimmed = path.trim();
          if (!trimmed) {
            this.pathChecking.set(false);
            return of({
              validation: null,
              suggestions: { base_path: '.', suggestions: [] } as PathSuggestionResponse,
            });
          }

          this.pathChecking.set(true);
          return forkJoin({
            validation: this.mangaService.validatePath(trimmed).pipe(
              catchError(() => of(null)),
            ),
            suggestions: this.mangaService.suggestPath(trimmed).pipe(
              catchError(() => of({ base_path: '.', suggestions: [] } as PathSuggestionResponse)),
            ),
          }).pipe(finalize(() => this.pathChecking.set(false)));
        }),
      )
      .subscribe(({ validation, suggestions }) => {
        this.pathValidation.set(validation);
        this.pathSuggestions.set(suggestions.suggestions ?? []);
        this.activeSuggestionIndex.set((suggestions.suggestions?.length ?? 0) > 0 ? 0 : -1);
        this.showSuggestions.set((suggestions.suggestions?.length ?? 0) > 0);
      });
  }

  ngOnDestroy(): void {
    this.pathInputSub?.unsubscribe();
  }

  onMangaSelected(manga: MangaSearchResult): void {
    this.selectedManga = manga;
    this.mangaPath = './mangas';
    this.pathValidation.set(null);
    this.pathInput$.next(this.mangaPath);
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
    this.pathValidation.set(null);
    this.pathSuggestions.set([]);
    this.showSuggestions.set(false);
    this.activeSuggestionIndex.set(-1);
  }

  onPathInput(value: string): void {
    this.mangaPath = value;
    this.error.set(null);
    this.pathValidation.set(null);
    this.activeSuggestionIndex.set(-1);
    this.pathInput$.next(value);
  }

  selectSuggestion(path: string): void {
    this.mangaPath = path;
    this.showSuggestions.set(false);
    this.activeSuggestionIndex.set(-1);
    this.pathInput$.next(path);
  }

  hideSuggestions(): void {
    setTimeout(() => this.showSuggestions.set(false), 120);
  }

  showPathSuggestions(): void {
    if (this.pathSuggestions().length > 0) {
      this.showSuggestions.set(true);
    }
  }

  onPathKeydown(event: KeyboardEvent): void {
    const suggestions = this.pathSuggestions();
    if (!this.showSuggestions() || suggestions.length === 0) {
      return;
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      this.activeSuggestionIndex.update((index) => (index + 1) % suggestions.length);
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault();
      this.activeSuggestionIndex.update((index) => (index <= 0 ? suggestions.length - 1 : index - 1));
      return
    }

    if (event.key === 'Enter') {
      const index = this.activeSuggestionIndex();
      if (index >= 0 && index < suggestions.length) {
        event.preventDefault();
        this.selectSuggestion(suggestions[index].path);
      }
      return
    }

    if (event.key === 'Escape') {
      this.showSuggestions.set(false);
      this.activeSuggestionIndex.set(-1);
    }
  }

  onConfirm(): void {
    if (!this.selectedManga) return;
    if (!this.mangaPath.trim()) {
      this.error.set('Please choose a destination path.');
      return;
    }
    if (this.pathValidation() && !this.pathValidation()!.valid) {
      this.error.set(this.pathValidation()!.message);
      return;
    }

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
    this.pathValidation.set(null);
    this.pathChecking.set(false);
    this.pathSuggestions.set([]);
    this.showSuggestions.set(false);
    this.activeSuggestionIndex.set(-1);
    this.mangaSelect?.clear();
  }

  close(): void {
    this.visible.set(false);
  }
}
