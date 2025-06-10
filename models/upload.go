package models

type PutObjectUpload struct {
	Acl             string `json:"acl"`
	ContentType     string `json:"content_type"`
	ContentEncoding string `json:"content_encoding"`
	Key             string `json:"key"`
	Platform        string `json:"platform"`
}
