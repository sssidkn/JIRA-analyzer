import { Component, OnInit } from '@angular/core';
import { ProjectService } from '../../data/services/project.service';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Project, MyProjectsResponse, StatisticResponse } from '../../data/models/project.model';

@Component({
  selector: 'app-comparison-page',
  imports: [
    CommonModule,
    FormsModule
  ],
  standalone: true,
  templateUrl: './comparison-page.component.html',
  styleUrl: './comparison-page.component.scss'
})
export class ComparisonPageComponent implements OnInit {
  projects: Project[] = [];
  selectedProjects: string[] = [];
  statistics: StatisticResponse['data'][] = [];
  currentPage = 1;
  itemsPerPage = 10;
  totalItems = 0;
  pageCount = 1;
  isLoading = false;
  showComparison = false;

  constructor(private projectService: ProjectService) {}

  ngOnInit(): void {
    this.loadProjects();
  }

  loadProjects(): void {
    this.isLoading = true;
    this.projectService.getDownloadedProjects(this.itemsPerPage, this.currentPage)
      .subscribe({
        next: (response: MyProjectsResponse) => {
          this.projects = response.data;
          this.totalItems = response.pageInfo.total;
          this.pageCount = response.pageInfo.pageCount;
          this.isLoading = false;
        },
        error: (error) => {
          console.error('Error loading projects:', error);
          this.isLoading = false;
        }
      });
  }

  analyzeSelectedProjects(): void {
    if (this.selectedProjects.length === 0) {
      alert('Please select at least one project');
      return;
    }

    this.isLoading = true;
    this.statistics = [];

    const statsPromises = this.selectedProjects.map(projectKey => {
      return this.projectService.getStatistic(projectKey).toPromise();
    });

    Promise.all(statsPromises)
      .then(results => {
        // Извлекаем data из каждого результата
        this.statistics = results
          .filter(r => r !== undefined && r.data !== undefined)
          .map(r => r!.data);
        this.showComparison = true;
        this.isLoading = false;
      })
      .catch(error => {
        console.error('Error fetching statistics:', error);
        this.isLoading = false;
      });
  }

  onPageChange(page: number): void {
    this.currentPage = page;
    this.loadProjects();
  }

  toggleProjectSelection(projectId: string): void {
    const index = this.selectedProjects.indexOf(projectId);
    if (index === -1) {
      this.selectedProjects.push(projectId);
    } else {
      this.selectedProjects.splice(index, 1);
    }
    this.showComparison = false;
  }
}
