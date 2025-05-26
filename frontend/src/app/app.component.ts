import {Component, inject} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {ProjectCardComponent} from './common-ui/project-card/project-card.component';
import {ProjectService} from './data/services/project.service';
import {JsonPipe} from '@angular/common';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, ProjectCardComponent, JsonPipe],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  projectService = inject(ProjectService)
  projects: any = []
  constructor() {
    this.projectService.getProjects()
      .subscribe(value => {
         this.projects = value
      })
  }
}
