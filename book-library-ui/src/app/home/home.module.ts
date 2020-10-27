import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {SidebarComponent} from './sidebar/sidebar.component';
import {HomeComponent} from './home.component';
import {LoginComponent} from '../login/login.component';
import {NavbarComponent} from './navbar/navbar.component';
import {SearchComponent} from './search/search.component';
import {LoanComponent} from './loan/loan.component';
import {BookComponent} from './book/book.component';

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [SidebarComponent, HomeComponent, LoginComponent, NavbarComponent, SidebarComponent, SearchComponent, LoanComponent, BookComponent]
})
export class HomeModule {
}
