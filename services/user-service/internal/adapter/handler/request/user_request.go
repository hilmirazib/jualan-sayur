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

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"email,required"`
}

type ResetPasswordRequest struct {
	Token                string `json:"token" validate:"required"`
	Password             string `json:"password" validate:"required,min=8"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
}

type UpdateProfileRequest struct {
	Email   string  `json:"email" validate:"required,email"`
	Name    string  `json:"name" validate:"required,min=2,max=100"`
	Phone   string  `json:"phone" validate:"required"`
	Address string  `json:"address" validate:"required"`
	Lat     float64 `json:"lat" validate:"required"`
	Lng     float64 `json:"lng" validate:"required"`
	Photo   string  `json:"photo" validate:"required"`
}
