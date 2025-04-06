package controllers

type Account struct {
	Provider string `json:"provider"`
	KeyID    string `json:"key_id"`
	Secret   string `json:"secret"`
}
