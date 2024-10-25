package utils

import (
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// QueryPaginatedList 通用分页函数
func QueryPaginatedList[T any](db *gorm.DB, page, pageSize int, model *T, queryCallback func(q *gorm.DB) *gorm.DB) ([]*T, int64, error) {
	var items []*T
	var total int64
	query := db.Model(model)
	if queryCallback != nil {
		query = queryCallback(query)
	}
	query = query.Count(&total)
	if pageSize != -1 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}
	err := query.Find(&items).Error
	return items, total, err
}

type IDType interface {
	GetID() string
}

func QueryListToMap[T IDType](db *gorm.DB, rawIdLst []string, t T, selectQuery string) map[string]T {
	idLst := lo.Uniq(rawIdLst)

	var items []T
	q := db.Where("id in ?", idLst)
	if selectQuery != "" {
		q = q.Select(selectQuery)
	}
	q.Find(&items)

	m := map[string]T{}
	for _, i := range items {
		m[i.GetID()] = i
	}

	return m
}

// QueryOneToManyMap K是列表对象，如Group，这里要查一个K的属性，如K.UserIds，然后T是这个对应的类型，如User
// flatten([x.UserIds for x in K]) -> Users
func QueryOneToManyMap[K IDType, T IDType](db *gorm.DB, items []K, f1 func(i K) []string, f2 func(i K, x []T), selectQuery string) {
	var arr []string
	for _, i := range items {
		arr = append(arr, f1(i)...)
	}

	var x T
	m := QueryListToMap(db, arr, x, selectQuery)

	for _, i := range items {
		var lst []T
		for _, id := range f1(i) {
			lst = append(lst, m[id])
		}
		f2(i, lst)
	}
}
