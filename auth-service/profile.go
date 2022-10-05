package authservice

import (
	"auth-service/m/v1/data"
	"encoding/json"
	"net/http"
)




func ProfileHandler(rw http.ResponseWriter, r *http.Request) {
	// validate the request first.
	if _, ok := r.Header["Email"]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Email Missing"))
		return
	}
	usr, exists := data.GetUserObject(r.Header["Email"][0])
	
	if !exists {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte("User Does not Exist"))
	}
	
   user, err  := json.Marshal(usr)
   if err !=nil {
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write([]byte("Internal Server Error"))
   }
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(user))
}