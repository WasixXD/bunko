import { Component } from '@angular/core';
import { ButtonDemo } from "@components/button-demo/button-demo"

@Component({
  selector: 'app-home-page',
  imports: [ButtonDemo],
  templateUrl: './home.page.html',
  styleUrl: './home.page.css',
})
export class HomePage {

}
