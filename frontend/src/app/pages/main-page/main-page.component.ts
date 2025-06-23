import {Component, inject, OnInit} from '@angular/core';
import {ProjectCardComponent} from '../../common-ui/project-card/project-card.component';
import {ProjectService} from '../../data/services/project.service';
import {PageInfo, Project, ProjectsResponse} from '../../data/models/project.model';
import {FormsModule} from "@angular/forms";
import {debounceTime, distinctUntilChanged, startWith, Subject, switchMap} from "rxjs";
import {ViewportScroller} from "@angular/common";

@Component({
    selector: 'app-main-page',
    standalone: true,
    imports: [
        ProjectCardComponent,
        FormsModule
    ],
    templateUrl: './main-page.component.html',
    styleUrl: './main-page.component.scss'
})

export class MainPageComponent implements OnInit {
    projectService = inject(ProjectService)
    projects: Project[] = []
    pageInfo: PageInfo = {
        pageCount: 0,
        projectsCount: 0,
        currentPage: 1
    };

    searchQuery = '';
    private searchSubject = new Subject<string>();
    isLoading = false;
    private readonly itemsPerPage = 30;
    private readonly maxVisiblePages = 5;
    visiblePages: number[] = [];

    constructor(private viewportScroller: ViewportScroller) {
    }

    ngOnInit(): void {
        this.setupSearch();
        this.searchSubject.next('');
    }

    onSearchChange(): void {
        this.searchSubject.next(this.searchQuery);
    }

    goToPage(page: number): void {
        if (page < 1 || page > this.pageInfo.pageCount || page === this.pageInfo.currentPage) {
            return;
        }
        this.viewportScroller.scrollToPosition([0, 0]);
        this.isLoading = true;
        this.projectService.getProjects(this.searchQuery, this.itemsPerPage, page)
            .subscribe({
                next: (projectsResponse) => {
                    this.projects = projectsResponse.projects;
                    this.pageInfo = projectsResponse.pageInfo;
                    this.isLoading = false;
                    this.pageInfo.currentPage = page;
                    this.visiblePages = this.updateVisiblePages();
                },
                error: (err) => {
                    console.error('Ошибка загрузки:', err);
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


    private setupSearch(): void {
        this.searchSubject.pipe(
            startWith(''),
            debounceTime(300),
            distinctUntilChanged(),
            switchMap(query => {
                this.isLoading = true;

                this.pageInfo.currentPage = 1;
                return this.projectService.getProjects(query, this.itemsPerPage, 1);
            })
        ).subscribe({
            next: (projectsResponse) => {
                this.projects = projectsResponse.projects;
                this.pageInfo = projectsResponse.pageInfo;
                this.isLoading = false;
                this.pageInfo.currentPage = 1;

                this.visiblePages = this.updateVisiblePages();

            },
            error: (err) => {
                console.error('Ошибка поиска:', err);
                this.isLoading = false;
            }
        });
    }
}
