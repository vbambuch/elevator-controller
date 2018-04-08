package common

import (
	"consts"
	"encoding/json"
	"helper"
)

func GetNotification(d interface{}) (consts.Notification) {
	data, err := json.Marshal(d)
	helper.HandleError(err, "JSON error")

	return data
}

func GetRawJSON(d interface{}) (json.RawMessage) {
	data, err := json.Marshal(d)
	helper.HandleError(err, "JSON raw error")

	return data
}

