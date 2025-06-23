import { Component, inject, OnDestroy } from '@angular/core';
import { FormsModule } from "@angular/forms";
import { ProjectCardComponent } from "../../common-ui/project-card/project-card.component";
import { ProjectService } from "../../data/services/project.service";
import { PageInfo, Project } from "../../data/models/project.model";
import { ViewportScroller } from "@angular/common";
import { Subject, takeUntil } from 'rxjs';

@Component({
    selector: 'app-my-projects-page',
    standalone: true,
    imports: [
        FormsModule,
        ProjectCardComponent,
    ],
    templateUrl: './my-projects-page.component.html',
    styleUrl: './my-projects-page.component.scss'
})
export class MyProjectsPageComponent implements OnDestroy {
    projectService = inject(ProjectService);
    projects: Project[] = [];
    pageInfo: PageInfo = {
        pageCount: 0,
        projectsCount: 0,
        currentPage: 1
    };

    isLoading = false;
    private readonly itemsPerPage = 30;
    private readonly maxVisiblePages = 5;
    visiblePages: (number)[] = [];

    private destroy$ = new Subject<void>();

    constructor(private viewportScroller: ViewportScroller) {}

    ngOnDestroy(): void {
        this.destroy$.next();
        this.destroy$.complete();
    }

    goToPage(page: number): void {
        if (page < 1 || page > this.pageInfo.pageCount || page === this.pageInfo.currentPage) {
            return;
        }

        this.viewportScroller.scrollToPosition([0, 0]);
        this.isLoading = true;

        this.projectService.getDownloadedProjects(this.itemsPerPage, page)
            .pipe(
                takeUntil(this.destroy$)
            )
            .subscribe({
                next: (projectsResponse) => {
                    this.projects = projectsResponse?.data || [];
                    this.pageInfo = {
                        currentPage: projectsResponse?.pageInfo?.currentPage || 1,
                        pageCount: projectsResponse?.pageInfo?.pageCount || 1,
                        projectsCount: projectsResponse?.pageInfo?.total || 0
                    };
                    this.visiblePages = this.updateVisiblePages();
                    this.isLoading = false;
                },
                error: (err) => {
                    console.error('Ошибка загрузки:', err);
                    this.projects = [];
                    this.isLoading = false;
                }
            });
    }

    updateVisiblePages(): number[] {
        if (this.pageInfo.pageCount <= 1) return [];
        const current = this.pageInfo.currentPage;
        const total = this.pageInfo.pageCount;
        const maxVisible = this.maxVisiblePages;

        if (total <= maxVisible) {
            return Array.from({length: total}, (_, i) => i + 1);
        }

        let start = Math.max(1, current - Math.floor(maxVisible / 2));
        let end = Math.min(total, start + maxVisible - 1);

        if (end - start + 1 < maxVisible) {
            start = Math.max(1, end - maxVisible + 1);
        }

        let pages: number[] = [];

        if (start > 1) {
            pages.push(1);
            if (start > 2) {
                pages.push(-1);
            }
        }

        for (let i = start; i <= end; i++) {
            pages.push(i);
        }

        if (end < total) {
            if (end < total - 1) {
                pages.push(-1);
            }
            pages.push(total);
        }

        return pages;
    }
}
