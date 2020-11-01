import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {RoutingPath} from './models/RoutingPath';
import {RouterModule, Routes} from '@angular/router';
import {LoginComponent} from './login/login.component';
import {AuthGuard} from './guards/authGuard';
import {GuestGuard} from './guards/guest.guard';

const routes: Routes = [
  {
    path: '',
    redirectTo: RoutingPath.LOGIN,
    pathMatch: 'full'
  },
  {
    path: RoutingPath.LOGIN,
    component: LoginComponent,
    canActivate: [GuestGuard]
  },
  {
    path: RoutingPath.BOOK_SEARCH,
    canActivate: [AuthGuard],
    loadChildren: () => import('./home/home.module').then((m) => m.HomeModule)
  },
  {
    path: '**',
    redirectTo: ''
  }
];

@NgModule({
  imports: [CommonModule, RouterModule.forRoot(routes, {enableTracing: false, paramsInheritanceStrategy: 'always'})],
  declarations: []
})
export class AppRoutingModule {
}
