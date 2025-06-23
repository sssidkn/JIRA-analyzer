import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {MyProjectsResponse, ProjectsResponse, StatisticResponse, Task1, Task2} from '../models/project.model';
import {Observable} from 'rxjs';

@Injectable({
    providedIn: 'root'
})
export class ProjectService {
    http = inject(HttpClient);

    getProjects(search: string, limit: number, page: number): Observable<ProjectsResponse> {
        return this.http.get<ProjectsResponse>(`/api/v1/connector/projects?limit=${limit}&page=${page}&search=${search}`);
    }

    updateProject(projectKey: string) {
        console.log('update project', projectKey);
        return this.http.post(`/api/v1/connector/updateProject`, {project_key: projectKey})
    };

    isDownloaded(id: string) {
        return this.http.get<{ status: string }>(`http://localhost:8080/api/v1/projects/${id}`);
    }

    getDownloadedProjects(limit: number, page: number): Observable<MyProjectsResponse> {
        return this.http.get<MyProjectsResponse>(`http://localhost:8080/api/v1/projects?limit=${limit}&page=${page}}`);
    }

    getStatistic(projectKey: string): Observable<StatisticResponse> {
        return this.http.get<StatisticResponse>(`http://localhost:8080/api/v1/projects/${projectKey}`);
    }

    makeGraphTask1(projectKey: string) {
        return this.http.post<Task1[]>(`http://localhost:8080/api/v1/graph/make/1?project=${projectKey}`, {});
    }

    getGraphTask1(projectKey: string) {
        return this.http.get<Task1[]>(`http://localhost:8080/api/v1/graph/make/1?project=${projectKey}`);
    }

    makeGraphTask2(projectKey: string) {
        return this.http.post<Task2[]>(`http://localhost:8080/api/v1/graph/make/1?project=${projectKey}`, {});
    }

    getGraphTask2(projectKey: string) {
        return this.http.get<Task2[]>(`http://localhost:8080/api/v1/graph/make/1?project=${projectKey}`);
    }
}
