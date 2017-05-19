package server

import (
	"ciphertalk/common/constants"
	"ciphertalk/server/auth"
	"ciphertalk/server/controller"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Initialize - registers routes and starts up the server
func Initialize() {
	router := mux.NewRouter()
	controller := controller.NewAPIController()
	registerRoutes(router, controller)

	var port = "3000"
	log.Println("Server started on port " + port)
	err := http.ListenAndServe(":"+port, handlers.LoggingHandler(os.Stdout, router))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func registerRoutes(router *mux.Router, controller *controller.APIController) {
	var handleWebsockets = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller.HandleWebsockets(w, r)
	})

	// route for sending and recieving messages
	router.Handle("/websockets", auth.JwtMiddleware.Handler(handleWebsockets)).Methods(constants.HTTPGet)

	// authentication route
	router.HandleFunc("/login", controller.Login).Methods(constants.HTTPPost)
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, From CipherTalk!. Your path is %q", html.EscapeString(r.URL.Path))
}
