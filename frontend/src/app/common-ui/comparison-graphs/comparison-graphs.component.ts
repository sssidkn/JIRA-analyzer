import {Component, Input, OnInit, SimpleChanges} from '@angular/core';
import {CommonModule} from '@angular/common';
import {ChartModule} from 'primeng/chart';
import {AnalyticsService} from '../../data/services/analytics.service';
import {MessageService} from 'primeng/api';
import {ToastModule} from 'primeng/toast';
import {ProgressSpinner} from 'primeng/progressspinner';

interface Task1 {
  count: number;
  time: string;
}

interface Task2 {
  count: number;
  priority: string;
}

@Component({
  selector: 'app-comparison-graphs',
  standalone: true,
  imports: [CommonModule, ChartModule, ToastModule, ProgressSpinner],
  templateUrl: './comparison-graphs.component.html',
  styleUrls: ['./comparison-graphs.component.scss'],
  providers: [MessageService]
})
export class ComparisonGraphsComponent implements OnInit {
  @Input() projectKeys: string[] = [];

  timeChartData: any;
  priorityChartData: any;

  chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      y: {beginAtZero: true}
    }
  };

  isLoading = false;
  showCharts = false;
  hasError = false;

  constructor(
      private analyticsService: AnalyticsService,
      private messageService: MessageService
  ) {}

  ngOnInit(): void {
    if (this.projectKeys?.length > 0) {
      this.loadGraphData();
    }
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['projectKeys'] && this.projectKeys?.length > 0) {
      this.loadGraphData();
    } else {
      this.showCharts = false;
    }
  }

  loadGraphData(): void {
    if (!this.projectKeys || this.projectKeys.length === 0) {
      this.showCharts = false;
      return;
    }

    this.isLoading = true;
    this.showCharts = false;
    this.hasError = false;

    const timeRequests = this.projectKeys.map(key =>
        this.analyticsService.makeGraphTask1(key).toPromise()
            .catch(error => {
              console.error(`Error loading time data for project ${key}:`, error);
              return null;
            })
    );

    const priorityRequests = this.projectKeys.map(key =>
        this.analyticsService.makeGraphTask2(key).toPromise()
            .then(response => response ? [response] : null) // Обернуть в массив
            .catch(error => {
              console.error(`Error loading priority data for project ${key}:`, error);
              return null;
            })
    );

    Promise.all([...timeRequests, ...priorityRequests])
        .then(results => {
          const timeResults = results.slice(0, this.projectKeys.length) as Task1[][];
          const priorityResults = results.slice(this.projectKeys.length) as Task2[][][]; // Теперь это массив массивов

          // Check if we have any valid data
          const hasValidTimeData = timeResults.some(r => r !== null);
          const hasValidPriorityData = priorityResults.some(r => r !== null);

          if (!hasValidTimeData && !hasValidPriorityData) {
            throw new Error('No valid data received for any project');
          }

          if (hasValidTimeData) {
            this.prepareTimeChartData(timeResults);
          }

          if (hasValidPriorityData) {
            const flattenedPriorityResults = priorityResults.map(r => r ? r[0] : null);
            this.preparePriorityChartData(flattenedPriorityResults);
          }

          this.showCharts = hasValidTimeData || hasValidPriorityData;
        })
        .catch(error => {
          console.error('Error loading graph data:', error);
          this.hasError = true;
          this.messageService.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Failed to load graph data'
          });
        })
        .finally(() => {
          this.isLoading = false;
        });
  }

  prepareTimeChartData(data: Task1[][]): void {
    const timeCategories = [
      '1 hour', '1-5 hours', '5-10 hours',
      '1-2 days', '2-5 days', '5+ days'
    ];

    const datasets = data.map((projectData, index) => {
      if (!projectData) {
        console.warn(`No time data for project ${this.projectKeys[index]}`);
        return {
          label: `Project ${this.projectKeys[index]}`,
          data: timeCategories.map(() => 0),
          backgroundColor: this.getColor(index, 0.7),
          borderColor: this.getColor(index, 1),
          borderWidth: 1
        };
      }

      const counts = timeCategories.map(category => {
        const item = projectData.find(d => d?.time === category);
        return item ? item.count : 0;
      });

      return {
        label: `Project ${this.projectKeys[index]}`,
        data: counts,
        backgroundColor: this.getColor(index, 0.7),
        borderColor: this.getColor(index, 1),
        borderWidth: 1
      };
    });

    this.timeChartData = {
      labels: timeCategories,
      datasets: datasets
    };
  }

  preparePriorityChartData(data: (Task2[] | null)[]): void {
    const priorityCategories = [
      'blocker', 'critical', 'major', 'minor', 'trivial'
    ];
    console.log('Priority results:', data);

    const datasets = data.map((projectData, index) => {
      if (!projectData) {
        console.warn(`No priority data for project ${this.projectKeys[index]}`);
        return {
          label: `Project ${this.projectKeys[index]}`,
          data: priorityCategories.map(() => 0),
          backgroundColor: this.getColor(index, 0.7),
          borderColor: this.getColor(index, 1),
          borderWidth: 1
        };
      }

      const counts = priorityCategories.map(priority => {
        const item = projectData.find(d => d?.priority === priority);
        return item ? item.count : 0;
      });

      return {
        label: `Project ${this.projectKeys[index]}`,
        data: counts,
        backgroundColor: this.getColor(index, 0.7),
        borderColor: this.getColor(index, 1),
        borderWidth: 1
      };
    });

    this.priorityChartData = {
      labels: priorityCategories,
      datasets: datasets
    };
  }

  private getColor(index: number, opacity: number): string {
    const colors = [
      `rgba(255, 99, 132, ${opacity})`,
      `rgba(54, 162, 235, ${opacity})`,
      `rgba(255, 206, 86, ${opacity})`,
      `rgba(75, 192, 192, ${opacity})`,
      `rgba(153, 102, 255, ${opacity})`,
      `rgba(255, 159, 64, ${opacity})`
    ];
    return colors[index % colors.length];
  }
}
