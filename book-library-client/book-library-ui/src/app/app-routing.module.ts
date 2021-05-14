import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';


const routes: Routes = [
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full'
  },
  {
    path: 'books',
    loadChildren: () => import('./book/book.module').then((m) => m.BookModule)
  },
  {
    path: 'users',
//    canActivate: [AuthGuard],
    loadChildren: () => import('./user/user.module').then((m) => m.UserModule)
  },
  {
    path: 'loans',
 //   canActivate: [AuthGuard],
    loadChildren: () => import('./loan/loan.module').then((m) => m.LoanModule)
  },
  {
    path: '**',
    redirectTo: ''
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
