package elevator

import (
	"consts"
	"encoding/json"
	"helper"
)

func GetNotification(notification interface{}) (consts.Notification) {

	//notification := NotificationData{e.floor, e.direction, e.cabQueue}
	data, err := json.Marshal(notification)
	helper.HandleError(err, "JSON error")

	return data
}

func GetRawJSON(d interface{}) json.RawMessage {
	data, err := json.Marshal(d)
	helper.HandleError(err, "JSON raw error")

	return data
}

