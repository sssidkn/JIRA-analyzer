import {Component, inject} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {ProjectCardComponent} from './common-ui/project-card/project-card.component';
import {ProjectService} from './data/services/project.service';
import {PageInfo, Project, ProjectsResponse} from "./data/models/project.model";

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, ProjectCardComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})

export class AppComponent {
  projectService = inject(ProjectService)
  projects: Project[] = []
  pageInfo: PageInfo | undefined;
  projectsResponse: ProjectsResponse | undefined;
  constructor() {
    this.projectService.getProjects("", 20, 1)
      .subscribe(projectsResponse => {
         this.projects = projectsResponse.projects;
         this.pageInfo = projectsResponse.pageInfo;
      })
  }
}
