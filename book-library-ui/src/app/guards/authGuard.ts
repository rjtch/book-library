import { Injectable } from '@angular/core';
import { CanActivate } from '@angular/router';
import {AuthServiceService} from '../services/auth-service.service';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate {
  constructor(private authService: AuthServiceService) {}

  canActivate(): boolean {
    if (!this.authService.isLoggedIn()) {
      this.authService.logout();
      return false;
    }
    return true;
  }
}
