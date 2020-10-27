import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import {Routes} from '@angular/router';
import {RoutingPath} from '../models/RoutingPath';

const routes: Routes = [
  {
    path: '',
    redirectTo: RoutingPath.LOGIN,
    pathMatch: 'full'
  },
  {
    path: RoutingPath.BOOK_API,
//    canActivate: [AuthGuard],
//    loadChildren: () => import('./home/home.module').then((m) => m.HomeModule)
  },
  {
    path: RoutingPath.BOOK_SEARCH,
//    canActivate: [AuthGuard],
//    loadChildren: () => import('./bank-search/bank-search.module').then((m) => m.BankSearchModule)
  },
  {
    path: '**',
    redirectTo: ''
  }
];

@NgModule({
  /*  imports: [RouterModule.forRoot(routes, { enableTracing: false, paramsInheritanceStrategy: 'always' })],
    exports: [RouterModule]*/
  imports: [
    CommonModule
  ],
  declarations: []
})
export class HomeRoutingModule { }
