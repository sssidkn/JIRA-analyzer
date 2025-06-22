import {Routes} from '@angular/router';
import {MainPageComponent} from './pages/main-page/main-page.component';
import {MyProjectsPageComponent} from './pages/my-projects-page/my-projects-page.component';
import {MenuComponent} from './common-ui/menu/menu.component';
import {LayoutComponent} from './common-ui/layout/layout.component';

export const routes: Routes = [
  {
    path: '', component: LayoutComponent, children: [
      {
        path: 'my_projects', component: MyProjectsPageComponent
      },
      {
        path: '', component: MainPageComponent
      }
    ]
  },
];
