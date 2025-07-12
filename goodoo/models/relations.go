package models

import (
	"reflect"
	"gorm.io/gorm"
)

// RelationManager handles relationship operations
type RelationManager struct {
	db *gorm.DB
}

// NewRelationManager creates a new relation manager
func NewRelationManager(db *gorm.DB) *RelationManager {
	return &RelationManager{db: db}
}

// LoadMany2One loads a many-to-one relationship
func (rm *RelationManager) LoadMany2One(recordID uint, modelName string, foreignKey string, targetModel interface{}) error {
	return rm.db.Where(foreignKey+" = ?", recordID).First(targetModel).Error
}

// LoadOne2Many loads a one-to-many relationship
func (rm *RelationManager) LoadOne2Many(recordID uint, modelName string, foreignKey string, targetSlice interface{}) error {
	return rm.db.Where(foreignKey+" = ?", recordID).Find(targetSlice).Error
}

// LoadMany2Many loads a many-to-many relationship
func (rm *RelationManager) LoadMany2Many(recordID uint, joinTable string, localKey string, foreignKey string, targetSlice interface{}) error {
	// First get the related IDs from the join table
	var relatedIDs []uint
	err := rm.db.Table(joinTable).
		Select(foreignKey).
		Where(localKey+" = ?", recordID).
		Pluck(foreignKey, &relatedIDs).Error
	
	if err != nil {
		return err
	}
	
	if len(relatedIDs) == 0 {
		return nil
	}
	
	// Then get the actual records
	return rm.db.Where("id IN ?", relatedIDs).Find(targetSlice).Error
}

// CreateMany2ManyLink creates a many-to-many relationship link
func (rm *RelationManager) CreateMany2ManyLink(recordID uint, relatedID uint, joinTable string, localKey string, foreignKey string) error {
	data := map[string]interface{}{
		localKey:   recordID,
		foreignKey: relatedID,
	}
	return rm.db.Table(joinTable).Create(data).Error
}

// DeleteMany2ManyLink removes a many-to-many relationship link
func (rm *RelationManager) DeleteMany2ManyLink(recordID uint, relatedID uint, joinTable string, localKey string, foreignKey string) error {
	return rm.db.Table(joinTable).
		Where(localKey+" = ? AND "+foreignKey+" = ?", recordID, relatedID).
		Delete(nil).Error
}

// UpdateOne2Many updates one-to-many relationships
func (rm *RelationManager) UpdateOne2Many(recordID uint, relatedIDs []uint, modelName string, foreignKey string) error {
	// First, unlink existing relationships
	err := rm.db.Model(&BaseModel{}).
		Where(foreignKey+" = ?", recordID).
		Update(foreignKey, nil).Error
	
	if err != nil {
		return err
	}
	
	// Then create new relationships
	if len(relatedIDs) > 0 {
		return rm.db.Model(&BaseModel{}).
			Where("id IN ?", relatedIDs).
			Update(foreignKey, recordID).Error
	}
	
	return nil
}

// RelationProxy provides lazy loading for relationships
type RelationProxy[T any] struct {
	loaded   bool
	value    T
	loader   func() (T, error)
	recordID uint
	rm       *RelationManager
}

// NewMany2OneProxy creates a new many-to-one proxy
func NewMany2OneProxy[T any](recordID uint, rm *RelationManager, modelName string, foreignKey string) *RelationProxy[T] {
	return &RelationProxy[T]{
		recordID: recordID,
		rm:       rm,
		loader: func() (T, error) {
			var result T
			err := rm.LoadMany2One(recordID, modelName, foreignKey, &result)
			return result, err
		},
	}
}

// NewOne2ManyProxy creates a new one-to-many proxy
func NewOne2ManyProxy[T any](recordID uint, rm *RelationManager, modelName string, foreignKey string) *RelationProxy[[]T] {
	return &RelationProxy[[]T]{
		recordID: recordID,
		rm:       rm,
		loader: func() ([]T, error) {
			var result []T
			err := rm.LoadOne2Many(recordID, modelName, foreignKey, &result)
			return result, err
		},
	}
}

// Get returns the value, loading it if necessary
func (rp *RelationProxy[T]) Get() (T, error) {
	if !rp.loaded {
		value, err := rp.loader()
		if err != nil {
			var zero T
			return zero, err
		}
		rp.value = value
		rp.loaded = true
	}
	return rp.value, nil
}

// Set sets the value
func (rp *RelationProxy[T]) Set(value T) {
	rp.value = value
	rp.loaded = true
}

// IsLoaded returns whether the value has been loaded
func (rp *RelationProxy[T]) IsLoaded() bool {
	return rp.loaded
}

// RelatedRecordSet provides operations on related records
type RelatedRecordSet[T any] struct {
	*RecordSet[T]
	parentID   uint
	foreignKey string
}

// NewRelatedRecordSet creates a new related recordset
func NewRelatedRecordSet[T any](db *gorm.DB, model T, parentID uint, foreignKey string) *RelatedRecordSet[T] {
	return &RelatedRecordSet[T]{
		RecordSet:  NewRecordSet(db, model),
		parentID:   parentID,
		foreignKey: foreignKey,
	}
}

// Add adds records to the relationship
func (rrs *RelatedRecordSet[T]) Add(records ...T) error {
	for _, record := range records {
		// Set the foreign key to link to parent
		recordValue := reflect.ValueOf(&record).Elem()
		foreignKeyField := recordValue.FieldByName(rrs.foreignKey)
		if foreignKeyField.IsValid() && foreignKeyField.CanSet() {
			foreignKeyField.SetUint(uint64(rrs.parentID))
		}
	}
	
	_, err := rrs.RecordSet.Create(records)
	return err
}

// Remove removes records from the relationship
func (rrs *RelatedRecordSet[T]) Remove(records ...T) error {
	for _, record := range records {
		// Unset the foreign key
		recordValue := reflect.ValueOf(&record).Elem()
		foreignKeyField := recordValue.FieldByName(rrs.foreignKey)
		if foreignKeyField.IsValid() && foreignKeyField.CanSet() {
			foreignKeyField.SetUint(0)
		}
	}
	
	// Update the records to remove the relationship
	var vals = map[string]interface{}{
		rrs.foreignKey: nil,
	}
	
	rs := &RecordSet[T]{
		db:      rrs.db,
		Records: records,
		model:   rrs.model,
	}
	
	return rs.Write(vals)
}