import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {MyProjectsResponse, ProjectsResponse, StatisticResponse, Task1, Task2} from '../models/project.model';
import {Observable} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  http = inject(HttpClient);

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
    return this.http.post<Task2[]>(`http://localhost:8080/api/v1/graph/make/2?project=${projectKey}`, {});
  }

  getGraphTask2(projectKey: string) {
    return this.http.get<Task2[]>(`http://localhost:8080/api/v1/graph/make/1?project=${projectKey}`);
  }
}
