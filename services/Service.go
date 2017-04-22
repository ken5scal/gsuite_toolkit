package services

import "net/http"

type Service interface {
	SetClient(client *http.Client) error
}