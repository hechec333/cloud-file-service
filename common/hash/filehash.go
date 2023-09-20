package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func FileHash(storeId, fileId int, fileName string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d%d%s", storeId, fileId, fileName)))
	return hex.EncodeToString(hash.Sum(nil))
}
