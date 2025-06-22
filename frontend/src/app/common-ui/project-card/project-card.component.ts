import { Component, Input, inject } from '@angular/core';
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
  isAnalyzed: boolean = false;
  isDownloaded: boolean = false;
  ngOnInit() {
    this.ps.isAnalyzed(this.project.key).subscribe({
      next: (analyzed) => this.isAnalyzed = analyzed,
      error: (err) => console.error('Error checking analysis status:', err)
    });

  }
}
