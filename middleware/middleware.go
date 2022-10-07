package middleware
import(
	"net/http"
	"auth-service/m/v1/jwt"
	"go.uber.org/zap"
	
)

 type TokenValidator struct {
	logger *zap.Logger
 }


 func NewTokenValidator (logger *zap.Logger) *TokenValidator{
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
		check, err := jwt.ValidateToken(token, jwt.GetSecret())

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("Token Validation Failed"))
			return
		}
		if !check {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Token Invalid"))
			return
		}
		next.ServeHTTP(rw, r)
	})
}