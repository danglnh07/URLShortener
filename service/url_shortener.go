package service

import "encoding/base64"

func GenerateID(url string) string {
	return base64.URLEncoding.EncodeToString([]byte(url))
}
