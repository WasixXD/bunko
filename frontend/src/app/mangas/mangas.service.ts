import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from 'app/environments/environment';
import { Observable } from 'rxjs';
import { Manga } from './mangas.model';

@Injectable({
  providedIn: 'root',
  
})
export class MangaService {
  
  constructor(private http: HttpClient) {}

  getMangas(): Observable<Manga[]> {
    return this.http.get<Manga[]>(`${environment.backendUrl}/mangas`)
  }

  getMangaById(id: number): Observable<Manga> {
    return this.http.get<Manga>(`${environment.backendUrl}/mangas/get/?id=${id}`);
  }

  deleteManga(id: number): Observable<unknown> {
    return this.http.delete(`${environment.backendUrl}/mangas/delete/?id=${id}`);
  }

  updateMetadata(id: number): Observable<Manga> {
    return this.http.post<Manga>(`${environment.backendUrl}/mangas/update-metadata/?id=${id}`, {});
  }

  validatePath(path: string): Observable<PathValidationResponse> {
    return this.http.post<PathValidationResponse>(`${environment.backendUrl}/validate/path`, { path });
  }

  suggestPath(path: string): Observable<PathSuggestionResponse> {
    return this.http.post<PathSuggestionResponse>(`${environment.backendUrl}/suggest/path`, { path });
  }
}

export interface PathValidationResponse {
  path: string;
  valid: boolean;
  exists: boolean;
  can_write: boolean;
  message: string;
}

export interface PathSuggestion {
  path: string;
  label: string;
}

export interface PathSuggestionResponse {
  base_path: string;
  suggestions: PathSuggestion[];
}
