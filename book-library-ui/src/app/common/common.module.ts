import { NgModule } from '@angular/core';
import { NavbarComponent } from './navbar/navbar.component';
import { SearchComponent } from './search/search.component';

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [NavbarComponent, SearchComponent]
})
export class CommonModule { }
