import { CommonModule } from '@angular/common';
import { Component, OnInit, signal } from '@angular/core';
import { Manga } from 'app/mangas/mangas.model';
import { MangaService } from 'app/mangas/mangas.service';
import { ChangeDetectorRef } from '@angular/core';


@Component({
  selector: 'app-home-page',
  imports: [CommonModule],
  standalone: true,
  templateUrl: './home.page.html',
  styleUrl: './home.page.css',
})
export class HomePage implements OnInit {
  mangas = signal<Manga[]>([])

  constructor(private mangaService: MangaService) {}

  ngOnInit() {
    this.mangaService.getMangas().subscribe({
      next: (data) => {
        this.mangas.set(data)
      },
      error: (err) => {
        console.error(err)
      }
    }) 
  }

}
