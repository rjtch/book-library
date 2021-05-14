export * from './books.service';
import { BooksService } from './books.service';
export * from './categories.service';
import { CategoriesService } from './categories.service';
export * from './loans.service';
import { LoansService } from './loans.service';
export * from './users.service';
import { UsersService } from './users.service';
export const APIS = [BooksService, CategoriesService, LoansService, UsersService];
