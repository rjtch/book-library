import {Injectable} from '@angular/core';
import {JwtHelperService} from '@auth0/angular-jwt';
import {HttpClient} from '@angular/common/http';
import {Router} from '@angular/router';
import {AutoLogoutService} from './auto-logout.service';
import {ShareDataService} from './share-data.service';
import {Observable, throwError} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {LoginRequest, UsersService} from '../../../../api';


@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private authTokenStorageKey = 'access_token';
  private jwtHelperService = new JwtHelperService();

  constructor(private http: HttpClient,
              private router: Router,
              private autoLogoutService: AutoLogoutService,
              private shareDataService: ShareDataService,
              private userService: UsersService) { }

  authorize(loginRequest: LoginRequest): Observable<string> {
    return this.userService.logInPost(loginRequest)
      .pipe(map(authResp => {
        // @ts-ignore
        return authResp.headers.get(this.authTokenStorageKey);
      }));
  }

  login(credential: any): Observable<boolean> {
    return this.authorize(credential).pipe(
      map((jwt) => {
        if (jwt !== undefined) {
          this.autoLogoutService.initializeTokenMonitoring();
          if (typeof jwt === "string") {
            sessionStorage.setItem(this.authTokenStorageKey, jwt);
          }
          return true;
        }
        return false;
      }),
      catchError((error) => {
        return throwError(error);
      })
    );
  }

  isLoggedIn(): boolean {
    return !this.jwtHelperService.isTokenExpired(this.getAuthorizationToken());
  }

  // tslint:disable-next-line:typedef
  logout() {
    this.autoLogoutService.resetMonitoringConfig();
    sessionStorage.removeItem(this.authTokenStorageKey);
    this.router.navigate(['/book']);
  }

  getAuthorizedUser(): string {
    return this.jwtHelperService.decodeToken(this.getAuthorizationToken()).name;
  }

  getAuthorizationToken(): string {
    return sessionStorage.getItem(this.authTokenStorageKey);
  }

  // tslint:disable-next-line:typedef
  setAuthToken(newToken: string) {
    sessionStorage.setItem(this.authTokenStorageKey, newToken);
  }
}
