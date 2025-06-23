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
  ) {
  }

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
    this.isLoading = true;
    this.showCharts = false;

    const timeRequests = this.projectKeys.map(key =>
      this.analyticsService.makeGraphTask1(key).toPromise()
    );

    const priorityRequests = this.projectKeys.map(key =>
      this.analyticsService.makeGraphTask2(key).toPromise()
    );

    Promise.all([...timeRequests, ...priorityRequests])
      .then(results => {
        const timeResults = results.slice(0, this.projectKeys.length) as Task1[][];
        const priorityResults = results.slice(this.projectKeys.length) as Task2[][];

        this.prepareTimeChartData(timeResults);
        this.preparePriorityChartData(priorityResults);

        this.showCharts = true;
        this.isLoading = false;
      })
      .catch(error => {
        console.error('Error loading graph data:', error);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Failed to load graph data'
        });
        this.isLoading = false;
      });
  }

  prepareTimeChartData(data: Task1[][]): void {
    const timeCategories = [
      '1 hour', '1-5 hours', '5-10 hours',
      '1-2 days', '2-5 days', '5+ days'
    ];

    const datasets = data.map((projectData, index) => {
      const counts = timeCategories.map(category => {
        const item = projectData.find(d => d.time === category);
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

  preparePriorityChartData(data: Task2[][]): void {
    const priorityCategories = [
      'blocker', 'critical', 'major', 'minor', 'trivial'
    ];

    const datasets = data.map((projectData, index) => {
      const counts = priorityCategories.map(priority => {
        const item = projectData.find(d => d.priority === priority);
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
