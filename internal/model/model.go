package model

type Identifiable interface {
	GetID() int
	SetID(id int)
}
