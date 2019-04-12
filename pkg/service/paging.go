package service

// FindOptions represents options passed to find methods
type FindOptions struct {
	Limit      int64
	Offset     int64
	SortBy     string
	Descending bool
}
