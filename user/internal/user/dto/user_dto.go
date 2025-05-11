package dto

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
type UserRequest struct {
	ID string `json:"id"`
}
