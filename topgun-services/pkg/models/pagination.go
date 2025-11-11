package models

type Pagination struct {
	Page    int   `query:"page"`
	PerPage int   `query:"per_page"`
	Total   int64 `query:"total" swaggerignore:"true"`
}

func (p *Pagination) GetPaginationString() string {
	return "page=" + string(rune(p.Page)) + "&per_page=" + string(rune(p.PerPage))
}
