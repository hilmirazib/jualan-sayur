package entity

type UserEntity struct {
	ID         int64
	Name       string
	Email      string
	Password   string
	RoleName   string
	RoleID     int64
	Address    string
	Lat        float64
	Lng        float64
	Phone      string
	Photo      string
	IsVerified bool
}
