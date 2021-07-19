package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/macsko/todolist/controllers"
	"github.com/macsko/todolist/database"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {
	// Getting the environmental variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	database.Init()

	r := mux.NewRouter()

	// Lists API routes
	listsApi := r.PathPrefix("/api/lists").Subrouter()
	listsApi.HandleFunc("", controllers.GetLists).Methods("GET")
	listsApi.HandleFunc("", controllers.CreateList).Methods("POST")
	listsApi.HandleFunc("/{listid}", controllers.UpdateList).Methods("PUT")
	listsApi.HandleFunc("/{listid}", controllers.DeleteList).Methods("DELETE")
	listsApi.HandleFunc("/{listid}", controllers.GetList).Methods("GET")
	listsApi.HandleFunc("/{listid}/tasks", controllers.GetTasks).Methods("GET")
	listsApi.HandleFunc("/{listid}/tasks/{taskid}", controllers.GetTask).Methods("GET")
	listsApi.HandleFunc("/{listid}/tasks", controllers.CreateTask).Methods("POST")
	listsApi.HandleFunc("/{listid}/tasks/{taskid}", controllers.UpdateTask).Methods("PUT")
	listsApi.HandleFunc("/{listid}/tasks/{taskid}", controllers.DeleteTask).Methods("DELETE")
	listsApi.Use(controllers.AuthenticationMiddleware)

	// Login API routes
	r.HandleFunc("/api/login", controllers.LoginUser).Methods("POST")
	r.HandleFunc("/api/register", controllers.RegisterUser).Methods("POST")
	r.HandleFunc("/api/logout", controllers.LogoutUser).Methods("GET")
	r.HandleFunc("/api/auth", controllers.AuthenticateUser).Methods("GET")

	// Accepting calls from frontend
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowCredentials: true,
	})
	handler := c.Handler(r)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
