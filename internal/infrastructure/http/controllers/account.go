package controllers

type Account struct {
	Provider  string `json:"provider"`
	Region    string `json:"region"`
	AccessKey string `json:"access_key"`
	Secret    string `json:"secret"`
}
