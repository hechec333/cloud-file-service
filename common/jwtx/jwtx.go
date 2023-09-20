package jwtx

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var TimeOutError string = "token expire timeout"

func IsTokenExpire(err error) bool {
	return err.Error() == TimeOutError
}

func GetToken(secretKey string, iat, seconds, uid int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["uid"] = uid
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func ParseToken(tokenString string, secretKey string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid := int64(claims["uid"].(float64))
		return uid, nil
	} else {
		return 0, err
	}
}

func GetAccessToken(secretKey string, seconds int64, data map[string]interface{}) (string, error) {
	claims := make(jwt.MapClaims)
	iat := time.Now().Unix()
	claims["iat"] = iat
	claims["exp"] = iat + seconds
	for k, v := range data {
		claims[k] = v
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func ParseAccessToken(secretKey string, tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	data := make(map[string]any)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		for k, v := range claims {
			if k == "exp" && v.(int64) > time.Now().Unix() {
				return data, errors.New(TimeOutError)
			}
			data[k] = v
		}
	}
	return data, err

}
