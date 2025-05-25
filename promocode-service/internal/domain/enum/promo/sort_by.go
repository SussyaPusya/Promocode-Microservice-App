package promo

type SortBy string

const (
	SortByCreatedAt   SortBy = "created_at"
	SortByActiveFrom  SortBy = "active_from"
	SortByActiveUntil SortBy = "active_until"
)
