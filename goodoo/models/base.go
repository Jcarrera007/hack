package models

import (
	"time"
	"gorm.io/gorm"
)

// BaseModel represents the common fields that all Odoo models have
type BaseModel struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreateUID uint           `gorm:"column:create_uid;index" json:"create_uid"`
	WriteUID  uint           `gorm:"column:write_uid;index" json:"write_uid"`
	CreateDate time.Time     `gorm:"column:create_date;autoCreateTime" json:"create_date"`
	WriteDate  time.Time     `gorm:"column:write_date;autoUpdateTime" json:"write_date"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Domain represents a search domain (filter conditions)
type Domain []interface{}

// RecordSet represents a collection of records with common operations
type RecordSet[T any] struct {
	db      *gorm.DB
	Records []T
	model   T
}

// NewRecordSet creates a new RecordSet
func NewRecordSet[T any](db *gorm.DB, model T) *RecordSet[T] {
	return &RecordSet[T]{
		db:      db,
		Records: make([]T, 0),
		model:   model,
	}
}

// Search finds records matching the given domain
func (rs *RecordSet[T]) Search(domain Domain, offset, limit int, order string) (*RecordSet[T], error) {
	var records []T
	query := rs.db.Model(&rs.model)
	
	// Apply domain conditions
	query = rs.applyDomain(query, domain)
	
	// Apply ordering
	if order != "" {
		query = query.Order(order)
	} else {
		query = query.Order("id")
	}
	
	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&records).Error
	if err != nil {
		return nil, err
	}
	
	return &RecordSet[T]{
		db:      rs.db,
		Records: records,
		model:   rs.model,
	}, nil
}

// Create creates one or more records
func (rs *RecordSet[T]) Create(vals []T) (*RecordSet[T], error) {
	err := rs.db.Create(&vals).Error
	if err != nil {
		return nil, err
	}
	
	return &RecordSet[T]{
		db:      rs.db,
		Records: vals,
		model:   rs.model,
	}, nil
}

// Read retrieves specified fields for records
func (rs *RecordSet[T]) Read(fields []string) ([]T, error) {
	var records []T
	query := rs.db.Model(&rs.model)
	
	if len(fields) > 0 {
		query = query.Select(fields)
	}
	
	// If we have specific records, read those
	if len(rs.Records) > 0 {
		var ids []uint
		for _, record := range rs.Records {
			// Use reflection to get ID field
			if idField, ok := any(record).(interface{ GetID() uint }); ok {
				ids = append(ids, idField.GetID())
			}
		}
		if len(ids) > 0 {
			query = query.Where("id IN ?", ids)
		}
	}
	
	err := query.Find(&records).Error
	return records, err
}

// Write updates records with given values
func (rs *RecordSet[T]) Write(vals map[string]interface{}) error {
	if len(rs.Records) == 0 {
		return nil
	}
	
	var ids []uint
	for _, record := range rs.Records {
		if idField, ok := any(record).(interface{ GetID() uint }); ok {
			ids = append(ids, idField.GetID())
		}
	}
	
	if len(ids) > 0 {
		return rs.db.Model(&rs.model).Where("id IN ?", ids).Updates(vals).Error
	}
	
	return nil
}

// Unlink deletes records
func (rs *RecordSet[T]) Unlink() error {
	if len(rs.Records) == 0 {
		return nil
	}
	
	var ids []uint
	for _, record := range rs.Records {
		if idField, ok := any(record).(interface{ GetID() uint }); ok {
			ids = append(ids, idField.GetID())
		}
	}
	
	if len(ids) > 0 {
		return rs.db.Where("id IN ?", ids).Delete(&rs.model).Error
	}
	
	return nil
}

// Count returns the number of records matching the domain
func (rs *RecordSet[T]) Count(domain Domain) (int64, error) {
	query := rs.db.Model(&rs.model)
	query = rs.applyDomain(query, domain)
	
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// applyDomain applies domain conditions to a GORM query
func (rs *RecordSet[T]) applyDomain(query *gorm.DB, domain Domain) *gorm.DB {
	// Simple domain implementation - in real Odoo this is much more complex
	// Domain format: [['field', 'operator', 'value'], ...]
	for _, condition := range domain {
		if condSlice, ok := condition.([]interface{}); ok && len(condSlice) == 3 {
			field := condSlice[0].(string)
			operator := condSlice[1].(string)
			value := condSlice[2]
			
			switch operator {
			case "=":
				query = query.Where(field+" = ?", value)
			case "!=":
				query = query.Where(field+" != ?", value)
			case ">":
				query = query.Where(field+" > ?", value)
			case ">=":
				query = query.Where(field+" >= ?", value)
			case "<":
				query = query.Where(field+" < ?", value)
			case "<=":
				query = query.Where(field+" <= ?", value)
			case "like":
				query = query.Where(field+" LIKE ?", value)
			case "ilike":
				query = query.Where(field+" ILIKE ?", value)
			case "in":
				query = query.Where(field+" IN ?", value)
			case "not in":
				query = query.Where(field+" NOT IN ?", value)
			}
		}
	}
	return query
}

// GetID returns the ID of the base model
func (bm *BaseModel) GetID() uint {
	return bm.ID
}