import { NgModule } from '@angular/core';
import { NavbarComponent } from './navbar/navbar.component';
import { SearchComponent } from './search/search.component';
import { BookComponent } from './book/book.component';
import { LoanComponent } from './loan/loan.component';

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [NavbarComponent, SearchComponent, BookComponent, LoanComponent]
})
export class CommonModule { }
