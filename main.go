package main

import (
	"auth-service/m/v1/auth-service"
	"auth-service/m/v1/monitormodule"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func PingHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Your App seems Healthy"))
}

func simplePostHandler(rw http.ResponseWriter, r *http.Request) {
	fileName, err := os.OpenFile("./metricDetails.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("error in file ops", zap.Error(err))
	}
	defer fileName.Close()

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error", zap.Error(err))
	}
	fileName.Write(resp)
	fileName.Write([]byte("\n"))
	rw.Write([]byte("Post Request Recieved for the Success"))
}

func main() {
	mainRouter := mux.NewRouter()
	authRouter := mainRouter.PathPrefix("/auth").Subrouter()

	log, _ := zap.NewProduction()
	defer log.Sync()

	err := godotenv.Load(".env")
	if err != nil {
		log.Error("Error loading .env file", zap.Error(err))
	}
	err = monitormodule.MonitorBinder(log)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Info("Starting...")

	suc := authservice.NewSigupController(log)
	sic := authservice.NewSiginController(log)



    // Main routes for pinging and testing prometheus
	mainRouter.HandleFunc("/ping", PingHandler)
	mainRouter.HandleFunc("/checkRoutine", simplePostHandler).Methods("POST")




    // authentication routes
	authRouter.HandleFunc("/signup", suc.SignupHandler)
	// The Signin will send the JWT back as we are making microservices.
	// The JWT token will make sure that other services are protected.
	// So, ultimately, we would need a middleware
	authRouter.HandleFunc("/signin", sic.SigninHandler)







    // CORS Header
    cors := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"http://localhost:3000"}))
	// Adding Prometheus http handler to expose the metrics
	// this will display our metrics as well as some standard metrics
	mainRouter.Path("/metrics").Handler(promhttp.Handler())
	// Add the middleware to different subrouter
	// HTTP server
	// Add time outs
	server := &http.Server{
		Addr:    ":9090",
		Handler:  cors(mainRouter),
	}
	fmt.Println("Listening on: 192.168.247.79:9090")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error Booting the Server")
	}


}
