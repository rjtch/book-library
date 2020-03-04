export class Book {
    image?: string;
    authors?: Array<string>;
    bookQuantity?: number;
    bookStatus?: 'LENT' | 'AVAILABLE' | 'UN_AVAILABLE' | 'UPLOADED' | 'UPDATED';
    description?: string;
    id?: string;
    category?: string;
    isbn?: string;
    lentDate?: string;
    modifyDate?: string;
    title?: string;
    uploadDate?: string;
  }
  