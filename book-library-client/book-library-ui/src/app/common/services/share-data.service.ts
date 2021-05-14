import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';


@Injectable({
  providedIn: 'root'
})
export class ShareDataService {
  private user = new BehaviorSubject(null);
  currentUser = this.user.asObservable();

  constructor() { }

  // tslint:disable-next-line:typedef
  updateUserDetails(body) {
    return null;
  }
}
