package main

import (
	"fmt"
	"net/http"

	"auth-service/m/v1/auth-service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	mainRouter := mux.NewRouter()
	authRouter := mainRouter.PathPrefix("/auth").Subrouter()
	log, _ := zap.NewDevelopment()
	defer log.Sync()
	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file", zap.Error(err))
	}

	// err = monitormodule.MonitorBinder(log)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	log.Info("Starting...")
	authRouter.HandleFunc("/signup", authservice.SignupHandler)
	// The Signin will send the JWT back as we are making microservices.
	// The JWT token will make sure that other services are protected.
	// So, ultimately, we would need a middleware
	// authRouter.HandleFunc("/signin", authservice.SigninHandler)
	// authRouter.HandleFunc("/profile")
	// Add the middleware to different subrouter
	// HTTP server
	// Add time outs
	server := &http.Server{
		Addr:    "localhost:9090",
		Handler: mainRouter,
	}
	fmt.Println("Listening on: 192.168.247.79:9090")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error Booting the Server")
	}
	fmt.Println("Here we are")
}
