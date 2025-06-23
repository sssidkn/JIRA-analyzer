import {Component} from '@angular/core';
import {RouterLink, RouterLinkActive} from '@angular/router';

@Component({
  selector: 'app-menu',
  templateUrl: './menu.component.html',
  imports: [
    RouterLink,
    RouterLinkActive
  ],
  styleUrl: './menu.component.scss'
})
export class MenuComponent {
  menuItems = [
    {
      label: 'All projects',
      link: '',
    },
    {
      label: 'My Projects',
      link: '/my_projects',
    },
    {
      label: 'Comparison',
      link: '/comparison',
    }
  ]
}
