import {
  Component,
  output,
  signal,
  inject,
  DestroyRef,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, debounceTime, distinctUntilChanged, switchMap, of } from 'rxjs';
import { environment } from 'app/environments/environment';

export interface MangaSearchResult {
  name: string;
  year: string;
  cover_path: string;
  status: string;
  url: string;
  provider: string;
}

@Component({
  selector: 'app-manga-search-select',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './search-select.html',
  styleUrls: ['./search-select.css'],
})
export class MangaSearchSelectComponent {
  mangaSelected = output<MangaSearchResult>();
  mangaCleared = output<void>();

  private http = inject(HttpClient);
  private destroyRef = inject(DestroyRef);
  private search$ = new Subject<string>();

  query = signal('');
  results = signal<MangaSearchResult[]>([]);
  selected = signal<MangaSearchResult | null>(null);
  loading = signal(false);
  open = signal(false);

  constructor() {
    this.search$
      .pipe(
        debounceTime(400),
        distinctUntilChanged(),
        switchMap((q) => {
          if (!q.trim()) {
            this.results.set([]);
            this.loading.set(false);
            return of(null);
          }
          this.loading.set(true);
          return this.http
            .get<Record<string, Omit<MangaSearchResult, 'provider'>[]>>(
              `${environment.backendUrl}/quick-search/manga?q=${encodeURIComponent(q)}`
            )
            .pipe(takeUntilDestroyed(this.destroyRef));
        })
      )
      .subscribe((res) => {
        if (!res) return;
        const flat: MangaSearchResult[] = [];
        for (const [provider, items] of Object.entries(res)) {
          items.forEach((item) => flat.push({ ...item, provider }));
        }
        this.results.set(flat);
        this.loading.set(false);
        this.open.set(true);
      });
  }

  onInput(value: string): void {
    this.query.set(value);
    this.search$.next(value);
  }

  select(item: MangaSearchResult): void {
    this.selected.set(item);
    this.query.set('');
    this.open.set(false);
    this.mangaSelected.emit(item);
  }

  clear(): void {
    this.selected.set(null);
    this.query.set('');
    this.results.set([]);
    this.open.set(false);
    this.mangaCleared.emit();
  }

  closeDropdown(): void {
    setTimeout(() => this.open.set(false), 150);
  }
}