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
}