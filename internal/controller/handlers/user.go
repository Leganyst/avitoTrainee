package controller

type UserRepository interface {
	GetUser()
	CreateUser()
	UpdateUser()
	DeleteUser()

	SetActive()
	GetReviews()
}
