package repository

type Identifiable interface {
	GetID() int
}

type Repository[T any] interface {
	FindByID(id int) (*T, error)
	Update(data *T) (*T, error)
	Insert(data *T) (*T, error)
	Delete(id int) error
}
