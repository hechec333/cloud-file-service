package url

import "github.com/golang-jwt/jwt"

func GetUrl(secretKey string, iat, seconds int64, fileId int) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["fid"] = fileId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func ParseUrl(tokenString string, secretKey string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid := int64(claims["fid"].(float64))
		return uid, nil
	} else {
		return 0, err
	}
}
