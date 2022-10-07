package authservice

import (
	"auth-service/m/v1/data"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)


var (
	signupRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_total",
		Help: "Total number of signup request",
	})
	signupSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_success",
		Help: "Successful signup requests",
	})
	signupFail = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_fail",
		Help: "Failed signup requests",
	})
	signupError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_error",
		Help: "Erroneous signup requests",
	})
)

type SignupController struct {
	logger           *zap.Logger
	promSignupTotal  prometheus.Counter
	promSigupSuccess prometheus.Counter
	promSignupFail   prometheus.Counter
	promSignupError  prometheus.Counter
}

func NewSigupController(logger *zap.Logger) *SignupController{
   return &SignupController{
	logger: logger,
	promSignupTotal: signupRequests,
	promSigupSuccess: signupSuccess,
	promSignupFail: signupFail,
	promSignupError: signupError,
   }
}

// adds the user to the database of users
func (ctrl *SignupController)  SignupHandler(rw http.ResponseWriter, r *http.Request) {
	// extra error handling should be done at server side to prevent malicious attacks
	if _, ok := r.Header["Email"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Email Missing"))
		return
	}
	if _, ok := r.Header["Username"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Username Missing"))
		return
	}
	if _, ok := r.Header["Passwordhash"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Passwordhash Missing"))
		return
	}
	if _, ok := r.Header["Fullname"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Fullname Missing"))
		return
	}

	// validate and then add the user
	check := data.AddUserObject(r.Header["Email"][0], r.Header["Username"][0], r.Header["Passwordhash"][0],
		r.Header["Fullname"][0], 0)
	// if false means username already exists
	if !check {
		rw.WriteHeader(http.StatusConflict)
		rw.Write([]byte("Email or Username already exists"))
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("User Created"))
}

