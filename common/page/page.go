package page

import "gorm.io/gorm"

const (
	FileFirst   = 0b0000
	FolderFirst = 0b0001
	TimeDesc    = 0b0010
)

var DefaultOrderOption = FolderFirst & TimeDesc

type PageHandler = func(*gorm.DB) *gorm.DB
type Pagger struct {
	Page      int
	PageSize  int
	PageOrder int
	OffSet    int
}

func DefaultOrderPageWarrper(p Pagger) PageHandler {
	return func(db *gorm.DB) *gorm.DB {
		offset := GetOffset(&p)
		return db.Offset(offset).Limit(p.PageSize).Order("CreateTime asc")
	}
}

func PageWarrpper(page int, pageSize int, desc int) {

}

func GetOffset(p *Pagger) int {
	if p.Page == 0 {
		p.Page = 1
	}
	switch {
	case p.PageSize > 20:
		p.PageSize = 20
	case p.PageSize <= 5:
		p.PageSize = 5
	}
	return (p.Page-1)*p.PageSize + p.OffSet
}
