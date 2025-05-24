package service

type Service interface {
	GetProjects() error
	GetProject(id int) error
	DeleteProject(id int) error
}
