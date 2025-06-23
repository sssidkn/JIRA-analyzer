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
