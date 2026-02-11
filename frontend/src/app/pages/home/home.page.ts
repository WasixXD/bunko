import { Component, OnInit } from '@angular/core';
import { ButtonDemo } from "@components/button-demo/button-demo"
import { Manga } from 'app/mangas/mangas.model';
import { MangaService } from 'app/mangas/mangas.service';

@Component({
  selector: 'app-home-page',
  imports: [ButtonDemo],
  templateUrl: './home.page.html',
  styleUrl: './home.page.css',
})
export class HomePage implements OnInit {
  mangas: Manga[] = []

  constructor(private mangaService: MangaService) { 
  }

  ngOnInit() {
    this.mangaService.getMangas().subscribe(data => {
      this.mangas = data
    }) 
  }


}
