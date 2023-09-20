package code_test

import (
	"log"
	"orm/common/code"
	"os"
	"testing"
)

func TestGenImage(t *testing.T) {

	key := code.GetRandStr(5)

	image := code.ImgText(500, 150, key)
	var file *os.File
	if _, err := os.Stat("captcha.png"); err != nil {
		if os.IsNotExist(err) {
			file, _ = os.Create("captcha.png")
		} else {
			panic(err)
		}
	} else {
		os.Remove("captcha.png")
		file, err = os.Create("captcha.png")
		if err != nil {
			panic(err)
		}
	}

	bytes, err := file.Write(image)
	if err != nil {
		log.Panicln(err)
	}

	log.Println(bytes)
}
