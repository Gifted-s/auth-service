package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	//"github.com/golang-jwt/jwt/v4"
)

const (
	CORRUPT_TOKEN = "Corrupt Token"
	INVALID_TOKEN = "Invalid Token"
	EXPIRED_TOKEN = "Expired Token"
)


type JWTClaim struct{
  Aud string
  Iss string
  Exp string
}

func GetSecret() string {
	return os.Getenv("JWT_SECRET")
}




// Function for generating the tokens.
func GenerateToken(header string, payload *JWTClaim, secret string) (string, error) {
	// create a new hash of type sha256. We pass the secret key to it
	h := hmac.New(sha256.New, []byte(secret))
	header64 := base64.StdEncoding.EncodeToString([]byte(header))
	// We then Marshal the payload which is a map. This converts it to a string of JSON.
	payloadstr, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error generating Token")
		return string(payloadstr), err
	}
	payload64 := base64.StdEncoding.EncodeToString(payloadstr)

	// Now add the encoded string.
	message := header64 + "." + payload64

	// We have the unsigned message ready.
	unsignedStr := header + string(payloadstr)

	// We write this to the SHA256 to hash it.
	h.Write([]byte(unsignedStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	//Finally we have the token
	tokenStr := message + "." + signature
	return tokenStr, nil
}

// This helps in validating the token
func ValidateToken(token string, secret string) error {
	// JWT has 3 parts separated by '.'
	splitToken := strings.Split(token, ".")
	// if length is not 3, we know that the token is corrupt
	if len(splitToken) != 3 {
		return errors.New(CORRUPT_TOKEN)
	}

	// decode the header and payload back to strings
	header, err := base64.StdEncoding.DecodeString(splitToken[0])
	if err != nil {
		return  err
	}
	payload, err := base64.StdEncoding.DecodeString(splitToken[1])
	if err != nil {
		return  err
	}
	//again create the signature
	unsignedStr := string(header) + string(payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(unsignedStr))

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	fmt.Println(signature)

	// if both the signature don???t match, this means token is wrong
	if signature != splitToken[2] {
	  return errors.New(INVALID_TOKEN)
	}
	var payloadMap JWTClaim
    json.Unmarshal(payload, &payloadMap)
	if payloadMap.Exp < fmt.Sprint(time.Now().Unix()){
		return errors.New(EXPIRED_TOKEN)
	}
	// This means the token matches
	return nil
}

