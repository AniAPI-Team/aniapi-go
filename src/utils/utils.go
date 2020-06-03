package utils

// PageInfo contains a specific page indexes information
type PageInfo struct {
	End    int
	Number int
	Start  int
	Size   int
}

var pageSize int = 10

// GetPageInfo returns a page start and end indexes
func GetPageInfo(page int) *PageInfo {
	if page < 1 {
		page = 1
	}

	return &PageInfo{
		End:    page * pageSize,
		Number: page,
		Start:  (page * pageSize) - pageSize,
		Size:   pageSize,
	}
}
