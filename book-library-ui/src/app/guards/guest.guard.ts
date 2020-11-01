import { Injectable } from '@angular/core';
import { CanActivate, Router } from '@angular/router';
import {AuthServiceService} from '../services/auth-service.service';

@Injectable({
  providedIn: 'root'
})
export class GuestGuard implements CanActivate {
  constructor(private authService: AuthServiceService, private router: Router) {}

  canActivate(): boolean {
//    const isLoggedIn = this.authService.isLoggedIn();
    const isLoggedIn = true;

    if (isLoggedIn) {
      this.router.navigate(['/search']);
    }
    return true;
  }
}
