import { Component } from '@angular/core';
import {NgbModalConfig} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'book-library-ui';

  constructor(ngbModalConfig: NgbModalConfig) {
    ngbModalConfig.backdrop = 'static';
    ngbModalConfig.centered = true;
    ngbModalConfig.keyboard = false;
  }
}
