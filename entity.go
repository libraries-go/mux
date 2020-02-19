package mux

import (
	"context"
	"fmt"
	"net/http"
)

//EntityOperation The entity operation
type EntityOperation int

const (
	//EntityOperationCreate create operation
	EntityOperationCreate = EntityOperation(iota)
	//EntityOperationUpdate create operation
	EntityOperationUpdate
	//EntityOperationDelete create operation
	EntityOperationDelete
	//EntityOperationGet create operation
	EntityOperationGet
	//EntityOperationGetList create operation
	EntityOperationGetList
)

//HTTPMethod returns the associated http.Method
func (op EntityOperation) HTTPMethod() string {
	switch op {
	case EntityOperationCreate:
		return http.MethodPost
	case EntityOperationUpdate:
		return http.MethodPut
	case EntityOperationDelete:
		return http.MethodDelete
	case EntityOperationGet:
	case EntityOperationGetList:
		return http.MethodGet
	}
	return ""
}

//BuildPath given a base path, this method will build a path that can be registered with mux
func (op EntityOperation) BuildPath(basepath string, prefix string) string {
	if prefix != "" {
		prefix = prefix + "/"
	}
	basepath = fmt.Sprint(prefix, basepath)
	if op == EntityOperationGet || op == EntityOperationUpdate || op == EntityOperationDelete {
		return fmt.Sprint(basepath, "/", "{id:[0-9]+}")
	}
	return basepath
}

type RouteDetail struct {
	req *http.Request
	// resp http.ResponseWriter		(will get the status code and the response bytes from the operation handler itself)
}

//Path returns the path of the route detail
func (r *RouteDetail) Path() string {
	return r.req.URL.Path
}

//Headers returns the incoming headers
func (r *RouteDetail) Headers() http.Header {
	return r.req.Header
}

//Context returns the incoming request context
func (r *RouteDetail) Context() context.Context {
	return r.req.Context()
}

//Body returns body from the incoming request
func (r *RouteDetail) Body() interface{} {
	return nil
}

type OperationHandler interface {
	HandleOperation(*RouteDetail) (int, []byte)
}

type HandleOperation func(*RouteDetail) (int, []byte)

func (h HandleOperation) HandleOperation(r *RouteDetail) (int, []byte) {
	fmt.Println("from handle operation", h)
	return h(r)
}

///////////////
type AuthenticateInput interface {
	Headers() http.Header
	Path() string
	Context() context.Context
}

type Authenticationhandler interface {
	Authenticate(AuthenticateInput) bool
}

type Authenticate func(i AuthenticateInput) bool

func (authHandler Authenticate) Authenticate(i AuthenticateInput) bool {
	return authHandler(i)
}

////////////////
type AuthorizationHandler interface {
	Authorize(AuthorizationInput) bool
}

type AuthorizationInput interface {
	Path() string
	Context() context.Context
	Body() interface{}
}

//This will be used internally to generate HEATEOAS
type Authorize func(i AuthorizationInput) bool

func (authHandler Authorize) Authorize(i AuthorizationInput) bool {
	return authHandler(i)
}

///////////////
type InputValidationHandler interface {
	Validate(ValidationInput) bool
}

type ValidationInput interface {
	Body() interface{}
}

//This will be used internally to generate HEATEOAS
type Validate func(i ValidationInput) bool

func (vHandler Validate) Validate(i ValidationInput) bool {
	return vHandler(i)
}

//Entity Stores information that will be used to construct automated Restful URLs, and handle them as easily as possible
type Entity struct {
	basePath          string
	operationHandlers map[EntityOperation]OperationHandler
	Authenticationhandler
	operationsForAuthentication []EntityOperation
	AuthorizationHandler
	InputValidationHandler
	children []*Entity
}

//NewEntity returns a new Entity
func NewEntity() *Entity {
	return &Entity{basePath: "/"}
}

//WithPath registers the base base path for the entity
func (e *Entity) WithPath(path string) *Entity {
	if e.basePath == "/" {
		e.basePath = path
	}
	return e
}

//HandleOperation registers the operation handler against for the given operation
func (e *Entity) HandleOperation(op EntityOperation, handler OperationHandler) *Entity {
	_, ok := e.operationHandlers[op]
	if !ok {
		e.operationHandlers[op] = handler
	}
	return e
}

//HandleOperationFunc registers the operation handler against for the given operation
func (e *Entity) HandleOperationFunc(op EntityOperation, f func(*RouteDetail) (int, []byte)) *Entity {
	_, ok := e.operationHandlers[op]
	fmt.Println(f)
	if !ok {
		if e.operationHandlers == nil {
			e.operationHandlers = make(map[EntityOperation]OperationHandler)
		}
		e.operationHandlers[op] = HandleOperation(f)
	}
	fmt.Println("inside handleoperationfunc", e.operationHandlers[op])
	fmt.Println("inside handleoperationfunc 2", e.operationHandlers[op].HandleOperation)
	return e
}

//AuthenticateFunc uses the given authentication handler for authentication
//By default, the authentication handler from router will be used
func (e *Entity) AuthenticateFunc(authFunc Authenticate, requiredOperations ...EntityOperation) *Entity {
	if e.Authenticationhandler == nil {
		e.Authenticationhandler = Authenticate(authFunc)
	}
	if e.operationsForAuthentication == nil {
		e.operationsForAuthentication = make([]EntityOperation, 5)
	}
	if len(requiredOperations) == 0 {
		//appending all operations in the list, if the populated operations are not allowed for entity, they will just be hanging
		e.operationsForAuthentication = append(e.operationsForAuthentication, EntityOperationCreate, EntityOperationUpdate, EntityOperationDelete, EntityOperationGet, EntityOperationGetList)
	} else {
		e.operationsForAuthentication = append(e.operationsForAuthentication, requiredOperations...)
	}
	return e
}

//Authticate uses the given authentication handler for authentication
//By default, the authentication handler from router will be used
func (e *Entity) Authticate(authHandler Authenticationhandler, requiredOperations ...EntityOperation) *Entity {
	return e.AuthenticateFunc(authHandler.Authenticate, requiredOperations...)
}

//AuthorizeFunc uses the given authorization handler func for authorization
func (e *Entity) AuthorizeFunc(authFunc Authorize) *Entity {
	if e.AuthorizationHandler == nil {
		e.AuthorizationHandler = Authorize(authFunc)
	}
	return e
}

//Authorize uses the given authorization handler for authorization
func (e *Entity) Authorize(authHandler AuthorizationHandler) *Entity {
	return e.AuthorizeFunc(authHandler.Authorize)
}

//ValidateFunc uses the given validation handler func for validation
func (e *Entity) ValidateFunc(v Validate) *Entity {
	if e.InputValidationHandler == nil {
		e.InputValidationHandler = Validate(v)
	}
	return e
}

//Validate uses the given validation handler for validation
func (e *Entity) Validate(vHandler InputValidationHandler) *Entity {
	return e.ValidateFunc(vHandler.Validate)
}

//ForChild adds Children to the to entity
func (e *Entity) ForChild(child *Entity, children ...*Entity) *Entity {
	e.children = append(e.children, child)
	if len(children) > 0 {
		e.children = append(e.children, children...)
	}
	return e
}
