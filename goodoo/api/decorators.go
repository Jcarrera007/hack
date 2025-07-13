package api

import (
	"context"
	"fmt"
	"reflect"

	"goodoo/http"
	"goodoo/logging"
	"goodoo/models"
)

// MethodType represents the API type of a method
type MethodType string

const (
	// ModelMethod - static method that operates on model class
	ModelMethod MethodType = "model"
	// RecordMethod - instance method that operates on record(s)
	RecordMethod MethodType = "record"
	// ModelCreateMethod - special create method
	ModelCreateMethod MethodType = "model_create"
	// PrivateMethod - method not exposed via RPC
	PrivateMethod MethodType = "private"
)

// APIMethod contains metadata about an API method
type APIMethod struct {
	Name         string            `json:"name"`
	Type         MethodType        `json:"type"`
	Public       bool              `json:"public"`
	Constrains   []string          `json:"constrains,omitempty"`
	Depends      []string          `json:"depends,omitempty"`
	OnChange     []string          `json:"onchange,omitempty"`
	Returns      string            `json:"returns,omitempty"`
	Groups       []string          `json:"groups,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Help         string            `json:"help,omitempty"`
	Handler      interface{}       `json:"-"`
	Model        *models.ModelDefinition `json:"-"`
	Logger       *logging.Logger   `json:"-"`
}

// APIRegistry manages API method registration and exposure
type APIRegistry struct {
	methods map[string]map[string]*APIMethod // model_name -> method_name -> method
	models  map[string]*models.ModelDefinition
	logger  *logging.Logger
}

// NewAPIRegistry creates a new API registry
func NewAPIRegistry() *APIRegistry {
	return &APIRegistry{
		methods: make(map[string]map[string]*APIMethod),
		models:  make(map[string]*models.ModelDefinition),
		logger:  logging.GetLogger("goodoo.api.registry"),
	}
}

// MethodBuilder helps create API methods with decorators
type MethodBuilder struct {
	method     *APIMethod
	registry   *APIRegistry
}

// NewMethod creates a new method builder
func (r *APIRegistry) NewMethod(modelName, methodName string, handler interface{}) *MethodBuilder {
	if _, exists := r.methods[modelName]; !exists {
		r.methods[modelName] = make(map[string]*APIMethod)
	}

	method := &APIMethod{
		Name:    methodName,
		Type:    RecordMethod, // default
		Public:  true,
		Handler: handler,
		Logger:  logging.GetLogger(fmt.Sprintf("goodoo.api.%s.%s", modelName, methodName)),
	}

	// Get model if registered
	if model, exists := models.GetFieldModel(modelName); exists {
		method.Model = model
	}

	r.methods[modelName][methodName] = method
	r.logger.Info("Registered API method: %s.%s", modelName, methodName)

	return &MethodBuilder{
		method:   method,
		registry: r,
	}
}

// Model decorator - marks method as model-level (static)
func (b *MethodBuilder) Model() *MethodBuilder {
	b.method.Type = ModelMethod
	return b
}

// Private decorator - marks method as private (not RPC accessible)
func (b *MethodBuilder) Private() *MethodBuilder {
	b.method.Type = PrivateMethod
	b.method.Public = false
	return b
}

// Constrains decorator - specifies constraint dependencies
func (b *MethodBuilder) Constrains(fields ...string) *MethodBuilder {
	b.method.Constrains = fields
	return b
}

// Depends decorator - specifies compute dependencies
func (b *MethodBuilder) Depends(fields ...string) *MethodBuilder {
	b.method.Depends = fields
	return b
}

// OnChange decorator - specifies onchange fields
func (b *MethodBuilder) OnChange(fields ...string) *MethodBuilder {
	b.method.OnChange = fields
	return b
}

// Returns decorator - specifies return model
func (b *MethodBuilder) Returns(modelName string) *MethodBuilder {
	b.method.Returns = modelName
	return b
}

// Groups decorator - specifies required user groups
func (b *MethodBuilder) Groups(groups ...string) *MethodBuilder {
	b.method.Groups = groups
	return b
}

// Context decorator - adds context variables
func (b *MethodBuilder) Context(ctx map[string]interface{}) *MethodBuilder {
	if b.method.Context == nil {
		b.method.Context = make(map[string]interface{})
	}
	for k, v := range ctx {
		b.method.Context[k] = v
	}
	return b
}

// Help sets help text for the method
func (b *MethodBuilder) Help(help string) *MethodBuilder {
	b.method.Help = help
	return b
}

// Register completes method registration
func (b *MethodBuilder) Register() *APIMethod {
	return b.method
}

// APICall represents a call to an API method
type APICall struct {
	ModelName string                 `json:"model"`
	Method    string                 `json:"method"`
	Args      []interface{}          `json:"args"`
	Kwargs    map[string]interface{} `json:"kwargs"`
	Context   map[string]interface{} `json:"context"`
	IDs       []int                  `json:"ids,omitempty"`
}

// APIResponse represents the response from an API call
type APIResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Warning string      `json:"warning,omitempty"`
}

// ExecuteCall executes an API method call
func (r *APIRegistry) ExecuteCall(ctx context.Context, call *APICall, req *http.Request) *APIResponse {
	// Get method
	modelMethods, exists := r.methods[call.ModelName]
	if !exists {
		return &APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Model '%s' not found", call.ModelName),
		}
	}

	method, exists := modelMethods[call.Method]
	if !exists {
		return &APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Method '%s' not found on model '%s'", call.Method, call.ModelName),
		}
	}

	// Check if method is public
	if !method.Public {
		return &APIResponse{
			Success: false,
			Error:   "Method is not accessible via RPC",
		}
	}

	// Check user permissions
	if err := r.checkPermissions(ctx, method, req); err != nil {
		return &APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Access denied: %v", err),
		}
	}

	// Prepare method context
	methodCtx := r.prepareContext(ctx, call, method)

	// Execute method based on type
	var result interface{}
	var err error

	switch method.Type {
	case ModelMethod:
		result, err = r.executeModelMethod(methodCtx, method, call)
	case ModelCreateMethod:
		result, err = r.executeCreateMethod(methodCtx, method, call)
	case RecordMethod:
		result, err = r.executeRecordMethod(methodCtx, method, call)
	default:
		err = fmt.Errorf("unknown method type: %s", method.Type)
	}

	if err != nil {
		method.Logger.ErrorCtx(ctx, "Method execution failed: %v", err)
		return &APIResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	method.Logger.InfoCtx(ctx, "Method executed successfully")
	return &APIResponse{
		Success: true,
		Result:  result,
	}
}

// checkPermissions validates user permissions for method access
func (r *APIRegistry) checkPermissions(ctx context.Context, method *APIMethod, req *http.Request) error {
	// Check user groups if specified
	if len(method.Groups) > 0 {
		// TODO: Implement user groups checking
		// For now, allow access if user is authenticated
		if req.GetUserID() == 0 {
			return fmt.Errorf("authentication required")
		}
		// userGroups := req.GetUserGroups() // TODO: Implement GetUserGroups method
		// hasAccess := false
		// for _, reqGroup := range method.Groups {
		// 	for _, userGroup := range userGroups {
		// 		if userGroup == reqGroup {
		// 			hasAccess = true
		// 			break
		// 		}
		// 	}
		// 	if hasAccess {
		// 		break
		// 	}
		// }
		// if !hasAccess {
		// 	return fmt.Errorf("user does not have required groups: %v", method.Groups)
		// }
	}

	return nil
}

// prepareContext prepares the execution context
func (r *APIRegistry) prepareContext(ctx context.Context, call *APICall, method *APIMethod) context.Context {
	// Add method context
	if method.Context != nil {
		for k, v := range method.Context {
			ctx = context.WithValue(ctx, k, v)
		}
	}

	// Add call context
	if call.Context != nil {
		for k, v := range call.Context {
			ctx = context.WithValue(ctx, k, v)
		}
	}

	return ctx
}

// executeModelMethod executes a model-level method
func (r *APIRegistry) executeModelMethod(ctx context.Context, method *APIMethod, call *APICall) (interface{}, error) {
	handler := reflect.ValueOf(method.Handler)
	if handler.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler is not a function")
	}

	// Prepare arguments
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(method.Model),
	}

	// Add call arguments
	for _, arg := range call.Args {
		args = append(args, reflect.ValueOf(arg))
	}

	// Call method
	results := handler.Call(args)
	
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}
	
	if len(results) > 0 {
		return results[0].Interface(), nil
	}
	
	return nil, nil
}

// executeRecordMethod executes a record-level method
func (r *APIRegistry) executeRecordMethod(ctx context.Context, method *APIMethod, call *APICall) (interface{}, error) {
	if len(call.IDs) == 0 {
		return nil, fmt.Errorf("record method requires IDs")
	}

	handler := reflect.ValueOf(method.Handler)
	if handler.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler is not a function")
	}

	// For now, we'll just pass the IDs and args
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(call.IDs),
	}

	for _, arg := range call.Args {
		args = append(args, reflect.ValueOf(arg))
	}

	results := handler.Call(args)
	
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}
	
	if len(results) > 0 {
		return results[0].Interface(), nil
	}
	
	return nil, nil
}

// executeCreateMethod executes a create method
func (r *APIRegistry) executeCreateMethod(ctx context.Context, method *APIMethod, call *APICall) (interface{}, error) {
	if len(call.Args) == 0 {
		return nil, fmt.Errorf("create method requires data")
	}

	// Validate data using model if available
	if method.Model != nil {
		for _, arg := range call.Args {
			if data, ok := arg.(map[string]interface{}); ok {
				if err := method.Model.ValidateData(data); err != nil {
					return nil, fmt.Errorf("validation failed: %w", err)
				}
			}
		}
	}

	return r.executeModelMethod(ctx, method, call)
}

// GetMethods returns all registered methods for a model
func (r *APIRegistry) GetMethods(modelName string) map[string]*APIMethod {
	return r.methods[modelName]
}

// GetAllMethods returns all registered methods
func (r *APIRegistry) GetAllMethods() map[string]map[string]*APIMethod {
	return r.methods
}

// GetPublicMethods returns only public methods for a model
func (r *APIRegistry) GetPublicMethods(modelName string) map[string]*APIMethod {
	methods := r.methods[modelName]
	if methods == nil {
		return nil
	}

	public := make(map[string]*APIMethod)
	for name, method := range methods {
		if method.Public {
			public[name] = method
		}
	}
	return public
}

// MethodInfo returns method information for API documentation
func (r *APIRegistry) GetMethodInfo(modelName, methodName string) map[string]interface{} {
	if methods, exists := r.methods[modelName]; exists {
		if method, exists := methods[methodName]; exists {
			info := map[string]interface{}{
				"name":       method.Name,
				"type":       string(method.Type),
				"public":     method.Public,
				"help":       method.Help,
				"constrains": method.Constrains,
				"depends":    method.Depends,
				"onchange":   method.OnChange,
				"returns":    method.Returns,
				"groups":     method.Groups,
				"context":    method.Context,
			}
			return info
		}
	}
	return nil
}

// Global API registry
var DefaultAPIRegistry = NewAPIRegistry()

// Convenience functions for using the default registry

func NewMethod(modelName, methodName string, handler interface{}) *MethodBuilder {
	return DefaultAPIRegistry.NewMethod(modelName, methodName, handler)
}

func ExecuteCall(ctx context.Context, call *APICall, req *http.Request) *APIResponse {
	return DefaultAPIRegistry.ExecuteCall(ctx, call, req)
}

func GetMethods(modelName string) map[string]*APIMethod {
	return DefaultAPIRegistry.GetMethods(modelName)
}

func GetPublicMethods(modelName string) map[string]*APIMethod {
	return DefaultAPIRegistry.GetPublicMethods(modelName)
}

func GetMethodInfo(modelName, methodName string) map[string]interface{} {
	return DefaultAPIRegistry.GetMethodInfo(modelName, methodName)
}