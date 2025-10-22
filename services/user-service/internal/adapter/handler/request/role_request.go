package request

type CreateRoleRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}
