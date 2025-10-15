package request

type SignInRequest struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateUserAccountRequest struct {
	Email                string `json:"email" validate:"email,required"`
	Name                 string `json:"name" validate:"required,min=2,max=100"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}
