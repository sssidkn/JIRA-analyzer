export interface Project {
    id: number;
    url: string;
    key: string;
    name: string;
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
