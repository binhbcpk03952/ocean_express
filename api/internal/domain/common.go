package domain

type PaginationParams struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func (p *PaginationParams) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 20
	}
	return (p.Page - 1) * p.Limit
}

func (p *PaginationParams) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	return p.Limit
}

func CalculateTotalPages(totalItems int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		pages++
	}
	return pages
}
