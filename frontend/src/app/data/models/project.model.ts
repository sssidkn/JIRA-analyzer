export interface Project {
    url: string;
    key: string;
    name: string;
    id: string;
}

export interface PageInfo {
    pageCount: string;
    projectsCount: string;
    currentPage: number;
}

export interface ProjectsResponse {
    projects: Project[];
    pageInfo: PageInfo;
}
