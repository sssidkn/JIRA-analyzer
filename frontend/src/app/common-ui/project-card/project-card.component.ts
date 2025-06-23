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

    isLoading: boolean = false;

    ngOnInit() {
        this.checkDownloadStatus();
    }

    checkDownloadStatus() {
        if (this.project?.id) {
            this.ps.isDownloaded(this.project.id).subscribe({
                next: (response) => {
                    this.isProjectDownloaded = true;
                },
                error: (err) => {
                    console.error(`Error checking download status ${this.project.key}:`, err);
                    this.isProjectDownloaded = false;
                }
            });
        }
    }

    updateProject() {
        this.isLoading = true;

        this.ps.updateProject(this.project.key).subscribe({
            next: () => {
                this.isProjectDownloaded = true;
                this.isLoading = false;
                this.ps.updateProject(this.project.id);
            },
            error: (err) => {
                console.error(`Error updating project ${this.project.key}:`, err);
                this.isLoading = false;
            }
        });
    }

}
