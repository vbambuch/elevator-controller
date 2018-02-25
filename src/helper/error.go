package helper

import (
	"fmt"
	"log"
)

func HandleError(err error, message string) {
	if err != nil {
		fmt.Println(message)
		log.Fatal(err)
	}

}
