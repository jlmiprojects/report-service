package profiles

import (
	"encoding/json"

	utils "blueassetgroup.com/reports-service/shared"
)

type LoginResponse struct {
	utils.Result
	ID          string           `json:"id"`
	Username    string           `json:"user_name"`
	Email       string           `json:"email"`
	PremiumMaps bool             `json:"premium_maps"`
	Settings    *ProfileSettings `json:"settings"`
	Reports     bool             `json:"reports"`
}

type LoginRequest struct {
	Username string `json:"user_name"`
	Password string `json:"password"`
}

func (p *LoginRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type LoginAsRequest struct {
	Username         string `json:"user_name"`
	Password         string `json:"password"`
	CustomerUsername string `json:"customer_user_name"`
}

func (p *LoginAsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type LoginAsResponse struct {
	utils.Result
	ID          string `json:"id"`
	Username    string `json:"user_name"`
	Email       string `json:"email"`
	PremiumMaps bool   `json:"premium_maps"`
	Token       string `json:"token"`
}

type ChangePasswordRequest struct {
	NewPassord  string `json:"new_password"`
	OldPassword string `json:"old_password"`
	ID          string `json:"id"` // This id is only used of we are changing the password for someone else, if I change my own password the JWT token is used
}

func (p *ChangePasswordRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type ChangePasswordResponse struct {
	utils.Result
}

/*type GetProfileImageRequest struct {
	ID string `json:"id"`
}

func (p *GetProfileImageRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type GetProfileImageResponse struct {
	utils.Result
	BucketName string `json:"bucket_name"`
	Name       string `json:"name`
}*/
