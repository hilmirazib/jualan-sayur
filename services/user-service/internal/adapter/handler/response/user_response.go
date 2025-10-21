package response

type SignInResponse struct {
	AccessToken string  `json:"access_token"`
	Role        string  `json:"role"`
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
}

type CreateUserAccountResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ProfileResponse struct {
	ID      int64   `json:"id"`
	Email   string  `json:"email"`
	Role    string  `json:"role"`
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Photo   string  `json:"photo"`
}

type ImageUploadResponse struct {
	ImageURL string `json:"image_url"`
}
