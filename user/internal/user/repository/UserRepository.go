package repository

func NewUserRepository() IUserRepository {
	return NewInMemoryUserRepo()
}
