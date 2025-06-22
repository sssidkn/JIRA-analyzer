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
}
