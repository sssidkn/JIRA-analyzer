import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class ProjectService {
  http = inject(HttpClient)

  baseApiUrl = 'http://localhost:8081/'

  getProjects() {
    return this.http.get(`${this.baseApiUrl}api/v1/connector/projects`)
  }
}
