package utils

import (
	"fmt"
	"strings"

	"topgun-services/pkg/models"

	"gorm.io/gorm"
)

func ApplySearch(db *gorm.DB, filter models.Search) *gorm.DB {
	if filter.Keyword == "" || filter.Column == "" {
		return db
	}
	columns := strings.Split(filter.Column, ",")
	var query string

	if len(columns) > 0 {
		var args []interface{}

		for _, column := range columns {
			if query != "" {
				query += " OR "
			}
			query += fmt.Sprintf("%s ILIKE ?", column)
			args = append(args, "%"+filter.Keyword+"%")
		}

		return db.Where(query, args...)
	} else {
		return db.Where(fmt.Sprintf("%s ILIKE ?", filter.Column), "%"+filter.Keyword+"%")
	}
}
func ApplyPagination(db *gorm.DB, pagination *models.Pagination, model interface{}) *gorm.DB {
	var total int64
	err := db.Model(model).Count(&total).Error
	if err != nil {
		return nil
	}

	pagination.Total = total
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PerPage < 1 || pagination.PerPage > 50 {
		pagination.PerPage = 10
	}

	return db.Offset((pagination.Page - 1) * pagination.PerPage)
}
