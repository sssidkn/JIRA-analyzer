import {Component, Input, inject} from '@angular/core';
import {Project} from '../../data/models/project.model';
import {ProjectService} from '../../data/services/project.service';

@Component({
  selector: 'app-project-card',
  imports: [],
  templateUrl: './project-card.component.html',
  styleUrl: './project-card.component.scss'
})
export class ProjectCardComponent {
  @Input() project!: Project
  ps = inject(ProjectService)
  isProjectDownloaded: boolean = false;

    ngOnInit() {
        if (this.project?.id) {
            this.ps.isDownloaded(this.project.id).subscribe({
                next: (response) => {
                    if (response?.status) {
                        this.isProjectDownloaded = response.status != "not exist";
                    }
                },
                error: (err) => {
                    console.error(`Error checking download status ${this.project.key}:`, err);
                    this.isProjectDownloaded = false;
                }
            });
        }
    }
}
