package models

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"` // có thể là object hoặc nil
}