package models

type Credential struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type HealthCheck struct {
	Status string `json:"healthCheck,omitempty"`
}
