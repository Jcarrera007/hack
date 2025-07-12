package models

import (
	"fmt"
	"reflect"
	"sync"
	"gorm.io/gorm"
	"goodoo/database"
)

// Environment represents the execution context (similar to Odoo's env)
type Environment struct {
	db       *gorm.DB
	user     uint
	dbName   string
	registry *ModelRegistry
}

// NewEnvironment creates a new environment
func NewEnvironment(db *gorm.DB, user uint) *Environment {
	return &Environment{
		db:       db,
		user:     user,
		registry: GetRegistry(),
	}
}

// NewEnvironmentForDB creates a new environment for a specific database
func NewEnvironmentForDB(dbName string, user uint) (*Environment, error) {
	db, err := database.GetDatabase(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database %s: %w", dbName, err)
	}
	
	return &Environment{
		db:       db,
		user:     user,
		dbName:   dbName,
		registry: GetRegistry(),
	}, nil
}

// GetDB returns the database connection
func (env *Environment) GetDB() *gorm.DB {
	return env.db
}

// GetUser returns the current user ID
func (env *Environment) GetUser() uint {
	return env.user
}

// ModelRegistry manages all registered models
type ModelRegistry struct {
	models map[string]reflect.Type
	env    *Environment
	mutex  sync.RWMutex
}

var registry *ModelRegistry
var once sync.Once

// GetRegistry returns the singleton model registry
func GetRegistry() *ModelRegistry {
	once.Do(func() {
		registry = &ModelRegistry{
			models: make(map[string]reflect.Type),
		}
	})
	return registry
}

// SetEnvironment sets the environment for the registry
func (r *ModelRegistry) SetEnvironment(env *Environment) {
	r.env = env
}

// Register registers a model with the registry
func (r *ModelRegistry) Register(name string, model interface{}) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	
	r.models[name] = modelType
}

// Get returns a new instance of the specified model
func (r *ModelRegistry) Get(name string) (interface{}, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	modelType, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model %s not found", name)
	}
	
	return reflect.New(modelType).Interface(), nil
}

// GetRecordSet returns a new RecordSet for the specified model
func (r *ModelRegistry) GetRecordSet(name string) (interface{}, error) {
	model, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	
	// Use reflection to create the appropriate RecordSet type
	recordSetType := reflect.TypeOf(&RecordSet[BaseModel]{})
	recordSetValue := reflect.New(recordSetType.Elem())
	
	// Set the db and model fields
	recordSetValue.Elem().FieldByName("db").Set(reflect.ValueOf(r.env.db))
	recordSetValue.Elem().FieldByName("model").Set(reflect.ValueOf(model).Elem())
	
	return recordSetValue.Interface(), nil
}

// ListModels returns all registered model names
func (r *ModelRegistry) ListModels() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.models {
		names = append(names, name)
	}
	return names
}

// Model is a helper function to get a RecordSet for a model
func Model[T any](env *Environment, model T) *RecordSet[T] {
	return NewRecordSet(env.db, model)
}