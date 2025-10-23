package entity

type PaginationEntity struct {
	Page      int `json:"page"`
	TotalCount int64 `json:"total_count"`
	PerPage   int `json:"per_page"`
	TotalPage int `json:"total_page"`
}
