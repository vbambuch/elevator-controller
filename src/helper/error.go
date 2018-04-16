package helper

import (
	"log"
	"consts"
)

func HandleError(err error, message string) {
	if err != nil {
		log.Println(consts.Red, message, consts.Neutral)
		log.Fatal(err)
	}

}
