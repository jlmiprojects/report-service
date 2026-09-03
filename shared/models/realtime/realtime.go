package realtime

import utils "blueassetgroup.com/reports-service/shared"

type RealTimeResponse struct {
	utils.Result `json:"result"`
	Message      *Message `json:"message"`
}
