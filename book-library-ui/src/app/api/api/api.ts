export * from './books.service';
import { BooksService } from './books.service';
export * from './loans.service';
import { LoansService } from './loans.service';
export * from './users.service';
import { UsersService } from './users.service';
export const APIS = [BooksService, LoansService, UsersService];
