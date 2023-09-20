package util

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func Uid() int32 {
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Int31()
	return randomInt
}

func Uuid() string {
	return uuid.New().String()
}
