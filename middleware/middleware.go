package middleware

import (
	"auth-service/m/v1/jwt"
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type TokenValidator struct {
	logger *zap.Logger
}

func NewTokenValidator(logger *zap.Logger) *TokenValidator {
	return &TokenValidator{
		logger: logger,
	}
}

// We want all our routes for REST to be authenticated. So, we validate the token
func (ctrl *TokenValidator) TokenValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// check if token is present
		if _, ok := r.Header["Token"]; !ok {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Token Missing"))
			return
		}
		token := r.Header["Token"][0]
		err := jwt.ValidateToken(token, jwt.GetSecret())
		if err != nil {
			errStr := fmt.Sprint(err)
			ctrl.logger.Error(errStr, zap.String("token",token))
			if errStr == jwt.CORRUPT_TOKEN || errStr == jwt.EXPIRED_TOKEN || errStr == jwt.INVALID_TOKEN {
				rw.WriteHeader(http.StatusUnauthorized)
			} else {
				rw.WriteHeader(http.StatusInternalServerError)
			}
			rw.Write([]byte(errStr))
			return
		}
		next.ServeHTTP(rw, r)
	})
}
