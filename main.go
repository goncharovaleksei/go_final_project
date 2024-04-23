package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func getPort() int {
	port := 7540
	envPort := os.Getenv("TODO_PORT")
	if len(envPort) > 0 {
		if eport, err := strconv.ParseInt(envPort, 10, 32); err == nil {
			port = int(eport)
		}
	}

	return port
}

func main() {
	DatabaseInit()

	r := chi.NewRouter()
	r.Mount("/", http.FileServer(http.Dir(webDir)))
	r.Get("/api/nextdate", NextDateReadGET)
	r.Post("/api/task", Auth(TaskAddPOST))
	r.Get("/api/tasks", Auth(TasksReadGET))
	r.Get("/api/task", Auth(TaskReadGET))
	r.Put("/api/task", Auth(TaskUpdatePUT))
	r.Post("/api/task/done", Auth(TaskDonePOST))
	r.Delete("/api/task", Auth(TaskDELETE))
	r.Post("/api/signin", SignInPOST)

	serverPort := getPort()
	log.Println(fmt.Sprintf("Server port: %d", serverPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}
