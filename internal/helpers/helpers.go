package helpers

import (
	"io/ioutil"
	"log"
)

func ConvertImgToByte(path string) []byte {
	// Read the entire file into a byte slice
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}
