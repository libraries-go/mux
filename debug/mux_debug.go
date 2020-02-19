package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/products", ProductsHandler).Headers("Content-Type", "application/json")
	r.HandleFunc("/articles", ArticlesHandler)

	// type a struct{}

	// var postH a
	// func (p a) HandleOperation(r *mux.RouteDetail) (int, []byte) {
	// 	return 200, []byte("entity post is handled")
	// }

	putH := func(r *mux.RouteDetail) (int, []byte) {
		fmt.Println("This is a new code block")
		return 200, []byte("entity PUT is handled")
	}
	e := mux.NewEntity().WithPath("/path").HandleOperationFunc(mux.EntityOperationCreate, func(r *mux.RouteDetail) (int, []byte) {
		return 200, []byte("entity post is handled")
	}).HandleOperationFunc(mux.EntityOperationUpdate, putH).AuthenticateFunc(func(i mux.AuthenticateInput) bool {
		return true
	}, mux.EntityOperationCreate).AuthorizeFunc(func(i mux.AuthorizationInput) bool {
		return true
	}).ValidateFunc(func(i mux.ValidationInput) bool {
		return true
	})
	r.HandleEntity(e)

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, this is product handler"))
}

func ArticlesHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, this is articles handler"))
}
