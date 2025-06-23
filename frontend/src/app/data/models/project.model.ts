export interface Project {
    url: string;
    key: string;
    name: string;
    id: string;
}

export interface PageInfo {
    pageCount: number;
    projectsCount: number;
    currentPage: number;
}

export interface ProjectsResponse {
    projects: Project[];
    pageInfo: PageInfo;
}

export interface MyProjectsPageInfo {
    pageCount: number;
    total: number;
    currentPage: number;
}
export interface MyProjectsResponse {
    data: Project[];
    pageInfo: MyProjectsPageInfo;
    _links?: any;
}

export interface StatisticResponse {
  data:
    {
      id: number,
      key: string,
      name: string,
      allIssuesCount: number,
      openedIssuesCount: number,
      closedIssuesCount: number,
      resolvedIssuesCount: number,
      reopenedIssuesCount: number,
      progressIssuesCount: number,
      averageTime: number,
      averageIssuesCount: number,
    }
}
