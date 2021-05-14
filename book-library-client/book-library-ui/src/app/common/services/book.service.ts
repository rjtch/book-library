import { Injectable } from '@angular/core';
import {BooksService} from '../../../../../api/books.service';
import {Observable} from 'rxjs';
import {BookInfo} from '../../../../../model/bookInfo';
import {BookInfoList} from '../../../../../model/bookInfoList';
import {UpdateBook} from '../../../../../model/updateBook';

@Injectable({
  providedIn: 'root'
})
export class BookService {

  constructor(private bookService: BooksService) { }

  public getBooksList(): Observable<BookInfoList> {
    return this.bookService.listAllBooks();
  }

  public  getBookByName(bookId: string): Observable<BookInfo> {
    return this.bookService.retreiveBook(bookId);
  }

  public UpdateBookById(bookId: string, updateBook: UpdateBook): Observable<BookInfo> {
    return this.bookService.updateBook(bookId, updateBook);
  }

  public deleteBookById(bookId: string): Observable<BookInfo> {
    return this.bookService.deleteBooks(bookId);
  }
}
