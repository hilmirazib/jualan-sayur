package response

type SignInResponse struct {
	AccessToken string `json:"access_token"`
	Role        string `json:"role"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
}

type CreateUserAccountResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ProfileResponse struct {
	ID      int64  `json:"id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Lat     string `json:"lat"`
	Lng     string `json:"lng"`
	Photo   string `json:"photo"`
}
