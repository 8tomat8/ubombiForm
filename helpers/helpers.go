package helpers

import (
	"log"
)

func Check(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}
