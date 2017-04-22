package actions

import "github.com/ken5scal/gsuite_toolkit/services"

type Action interface {
	SetService(service services.Service) error
}
