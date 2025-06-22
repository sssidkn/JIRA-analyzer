import {Component, inject} from '@angular/core';
import {ProjectCardComponent} from '../../common-ui/project-card/project-card.component';
import {ProjectService} from '../../data/services/project.service';
import {PageInfo, Project, ProjectsResponse} from '../../data/models/project.model';

@Component({
  selector: 'app-main-page',
  imports: [
    ProjectCardComponent
  ],
  templateUrl: './main-page.component.html',
  styleUrl: './main-page.component.scss'
})
export class MainPageComponent {
  projectService = inject(ProjectService)
  projects: Project[] = []
  pageInfo: PageInfo | undefined;
  projectsResponse: ProjectsResponse | undefined;
  constructor() {
    this.projectService.getProjects("AT", 20, 1)
      .subscribe(projectsResponse => {
        console.log('Полный ответ:', projectsResponse);
        this.projects = projectsResponse.projects;
        this.pageInfo = projectsResponse.pageInfo;
      })
  }
}
