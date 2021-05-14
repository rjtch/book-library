import {NgModule} from '@angular/core';
import { NavbarComponent } from './navbar/navbar.component';
import { BookComponent } from './book/book.component';
import { FooterComponent } from './footer/footer.component';
import {ReactiveFormsModule} from '@angular/forms';


@NgModule({
  declarations: [
    NavbarComponent,
    BookComponent,
    FooterComponent
  ],
  exports: [
    NavbarComponent,
    FooterComponent
  ],
    imports: [
        ReactiveFormsModule
    ]
})
export class SharedModule { }
