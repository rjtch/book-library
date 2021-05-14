import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class StorageService {

  constructor() { }

  public getUserName(): string {
    return localStorage.getItem(Store.USERNAME);
  }

  public setUserName(userName: string): void {
    localStorage.setItem(Store.USERNAME, userName);
  }

  public clearStorage() {
    localStorage.clear();
  }

  // public isLoggedIn(): boolean {
  //   return this.isAnySessionValid();
  // }

}

enum Store {
  USERNAME = 'USERNAME',
  XSRF_TOKEN = 'XSRF_TOKEN',
  MAX_VALID_UNTIL = 'MAX_VALID_UNTIL_TIMESTAMP'
}
