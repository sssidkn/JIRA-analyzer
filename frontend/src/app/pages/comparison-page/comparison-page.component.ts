import {Component, OnInit} from '@angular/core';
import {ProjectService} from '../../data/services/project.service';
import {CommonModule} from '@angular/common';
import {FormsModule} from '@angular/forms';
import {Project, MyProjectsResponse, StatisticResponse} from '../../data/models/project.model';
import {ComparisonGraphsComponent} from "../../common-ui/comparison-graphs/comparison-graphs.component";
import {ChartModule} from "primeng/chart";
import {ToastModule} from "primeng/toast";
import {ProgressSpinnerModule} from "primeng/progressspinner";

@Component({
    selector: 'app-comparison-page',
    imports: [
        CommonModule,
        FormsModule,
        ComparisonGraphsComponent,
        ChartModule,
        ToastModule,
        ProgressSpinnerModule
    ],
    standalone: true,
    templateUrl: './comparison-page.component.html',
    styleUrl: './comparison-page.component.scss'
})
export class ComparisonPageComponent implements OnInit {
    projects: Project[] = [];
    selectedProjects: string[] = [];
    selectedProjectsForGraph: string[] = [];
    statistics: StatisticResponse['data'][] = [];
    currentPage = 1;
    itemsPerPage = 10;
    totalItems = 0;
    pageCount = 1;
    isLoading = false;
    showComparison = false;
    showGraphs = false;

    constructor(private projectService: ProjectService) {
    }

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
        this.showGraphs = false;

        const statsPromises = this.selectedProjects.map(projectId => {
            return this.projectService.getStatistic(projectId).toPromise();
        });

        Promise.all(statsPromises)
            .then(results => {
                this.statistics = results
                    .filter(r => r !== undefined && r.data !== undefined)
                    .map(r => r!.data);
                this.showComparison = true;
                this.showGraphs = true;
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

    toggleProjectSelection(project: Project): void {
        const idIndex = this.selectedProjects.indexOf(project.id);
        if (idIndex === -1) {
            this.selectedProjects.push(project.id);
            this.selectedProjectsForGraph.push(project.key);
        } else {
            this.selectedProjects.splice(idIndex, 1);
            this.selectedProjectsForGraph.splice(idIndex, 1);
        }
        this.showComparison = false;
    }
}
