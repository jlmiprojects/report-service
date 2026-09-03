package uploads

import utils "blueassetgroup.com/reports-service/shared"

type Upload struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type UploadResult struct {
	utils.Result
	Uploads []Upload `json:"uploads"`
}
