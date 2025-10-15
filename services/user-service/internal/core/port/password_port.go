package port

type PasswordInterface interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}
