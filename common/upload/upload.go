package upload

import "github.com/golang-jwt/jwt"

func GetUploadToken(secretKey string, iat, seconds int64, uploadId string, fileId int) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["uploadId"] = uploadId
	claims["fileId"] = fileId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func ParseUploadToken(tokenString string, secretKey string) (string, int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uploadId := (claims["uploadId"].(string))
		fileId := (claims["fileId"].(int))
		return uploadId, fileId, nil
	} else {
		return "", 0, err
	}
}
