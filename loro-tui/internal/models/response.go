package models

type LoginResponse struct {
	Username string `json:"username,omitempty"`
	Token    string `json:"token,omitempty"`
}
