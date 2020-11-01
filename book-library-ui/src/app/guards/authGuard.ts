import { Injectable } from '@angular/core';
import { CanActivate } from '@angular/router';
import {AuthServiceService} from '../services/auth-service.service';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate {
  constructor(private authService: AuthServiceService) {}

  canActivate(): boolean {
    return true;
  }
}
