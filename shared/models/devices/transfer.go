package devices

import (
	"encoding/json"
	"time"
)

const TransferStatusDone = "DONE"
const TransferStatusBusy = "BUSY"
const TransferStatusPending = "PENDING"
const TransferStatusFailed = "FAILED"

type Transfer struct {
	FromIMEI    string    `json:"from_imei"`
	ToIMEI      string    `json:"to_imei"`
	Status      string    `json:"status"`
	DateStarted time.Time `json:"date_start"`
	DateEnded   time.Time `json:"date_end"`
	User        string    `json:"string"`
	Message     string    `json:"message"`
}

type TransferRequest struct {
	FromIMEI string `json:"from_imei"`
	ToIMEI   string `json:"to_imei"`
}

func (u *TransferRequest) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)

	if err != nil {
		return nil
	}

	return nil
}
