package authservice

import (
	"auth-service/m/v1/data"
	"auth-service/m/v1/jwt"
	"errors"
	"fmt"
	"net/http"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	signinRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_total",
		Help: "Total number of signin request",
	})
	signinSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_success",
		Help: "Successful signin requests",
	})
	signinFail = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_fail",
		Help: "Failed signin requests",
	})
	signinError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signin_error",
		Help: "Erroneous signin requests",
	})
)

type SignInController struct {
	logger           *zap.Logger
	promSigninTotal  prometheus.Counter
	promSiginSuccess prometheus.Counter
	promSigninFail   prometheus.Counter
	promSigninError  prometheus.Counter
}

func NewSiginController(logger *zap.Logger) *SignInController{
   return &SignInController{
	logger: logger,
	promSigninTotal: signinRequests,
	promSiginSuccess: signinSuccess,
	promSigninFail: signinFail,
	promSigninError: signinError,
   }
}

// we need this function to be private
func getSignedToken() (string, error) {
	// we make a JWT Token here with signing method of ES256 and claims.
	// claims are attributes.
	// aud - audience
	// iss - issuer
	// exp - expiration of the Token
	claimsMap := map[string]string{
		"aud": "frontend.knowsearch.ml",
		"iss": "knowsearch.ml",
		"exp": fmt.Sprint(time.Now().Add(time.Minute * 1).Unix()),
	}
	// here we provide the shared secret. It should be very complex.
	// Also, it should be passed as a System Environment variable

	secret := "Secure_Random_String"
	header := "HS256"
	tokenString, err := jwt.GenerateToken(header, claimsMap, secret)
	if err != nil {
		return tokenString, err
	}
	return tokenString, nil
}

// searches the user in the database.
func validateUser(email string, passwordHash string) (bool, error) {
	usr, exists := data.GetUserObject(email)
	if !exists {
		return false, errors.New("user does not exist")
	}
	passwordCheck := usr.ValidatePasswordHash(passwordHash)

	if !passwordCheck {
		return false, nil
	}
	return true, nil
}

func (ctrl *SignInController) SigninHandler(rw http.ResponseWriter, r *http.Request) {
	ctrl.promSigninTotal.Inc()
	// validate the request first.
	if _, ok := r.Header["Email"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Email Missing"))
		ctrl.promSigninFail.Inc()
		return
	}
	if _, ok := r.Header["Passwordhash"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Passwordhash Missing"))
		ctrl.promSigninFail.Inc()
		return
	}
	// let’s see if the user exists
	valid, err := validateUser(r.Header["Email"][0], r.Header["Passwordhash"][0])
	if err != nil {
		// this means either the user does not exist
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte("User Does not Exist"))
		ctrl.promSigninFail.Inc()
		return
	}

	if !valid {
		// this means the password is wrong
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte("Incorrect Password"))
		ctrl.promSigninFail.Inc()
		return
	}
	tokenString, err := getSignedToken()
	if err != nil {
		fmt.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Internal Server Error"))
		ctrl.promSigninError.Inc()
		return
	}
    
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(tokenString))
	ctrl.promSiginSuccess.Inc()
}
