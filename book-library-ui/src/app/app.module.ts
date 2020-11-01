import {BrowserModule} from '@angular/platform-browser';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule} from '@angular/core';

import {AppComponent} from './app.component';
import {LoginComponent} from './login/login.component';
import {NavbarComponent} from './home/navbar/navbar.component';
import {CommonModule} from '@angular/common';
import {AppRoutingModule} from './app-routing.module';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';


/*export function apiConfigFactory(): Configuration {
  const params: ConfigurationParameters = {
    basePath: environment.BOOK_LIBRARY_API,
    withCredentials: true
  };

  return new Configuration(params);
}*/


@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    NavbarComponent,
  ],
  imports: [
    AppRoutingModule,
    BrowserModule,
    BrowserAnimationsModule,
    BrowserModule,
    CommonModule,
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule {
}
