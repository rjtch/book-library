import {StorageService} from '../services/storage.service';
import {AuthService} from '../services/auth.service';
import {BooksService} from '../../../../api';
import { Component, EventEmitter, Output } from '@angular/core';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { FormControl } from '@angular/forms';

@Component({
  selector: 'app-navbar',
  templateUrl: './navbar.component.html',
  styleUrls: ['./navbar.component.scss']
})
export class NavbarComponent {
  title = 'book-library';
  @Output() keyword = new EventEmitter();
  searchTerm = new FormControl();

  constructor(private storageService: StorageService,
              private authService: AuthService,
              private bookService: BooksService) {
    this.searchTerm.valueChanges.pipe(debounceTime(500), distinctUntilChanged()).subscribe(inputData => {
      this.keyword.emit(inputData);
    });
  }

  onLogout() {
    this.authService.logout();
    this.storageService.clearStorage();
  }

  getBookByName(bookName: string) {
    return this.bookService.retreiveBook(bookName);
  }

  isLoggedIn(): boolean {
    return this.authService.isLoggedIn();
  }

  getUserName(): string {
    return this.storageService.getUserName();
  }

}
