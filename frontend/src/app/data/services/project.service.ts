import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {ProjectsResponse} from '../models/project.model';
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
    return this.http.post(`/api/v1/connector/updateProject`, `{"project_key": ${projectKey}}`);
  }

  isAnalyzed(projectKey: string) {
    return this.http.get<boolean>(`/api/v1/isAnalyzed?project=${projectKey}`);
  }

  isDownloaded(id: string) {
    return this.http.get<string>(`/api/v1/projects/${id}`);
  }

}
