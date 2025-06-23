import { Component } from '@angular/core';
import {FormsModule} from '@angular/forms';
import {ProjectService} from '../../data/services/project.service';
import {debounceTime, distinctUntilChanged, Subject, switchMap} from 'rxjs';

@Component({
  selector: 'app-search-line',
  imports: [
    FormsModule
  ],
  templateUrl: './search-line.component.html',
  styleUrl: './search-line.component.scss'
})

export class SearchLineComponent {
  searchQuery = '';
  private searchSubject = new Subject<string>();

  constructor(private apiService: ProjectService) {
    this.setupSearch();
  }

  onSearchChange(): void {
    this.searchSubject.next(this.searchQuery);
  }

  private setupSearch(): void {
    this.searchSubject.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      switchMap(query =>
        this.apiService.getProjects(query, 30, 1)
      )
    ).subscribe({
      next: (results) => {
        console.log('Результаты:', results);
      },
      error: (err) => {
        console.error('Ошибка поиска:', err);
      }
    });
  }
}
