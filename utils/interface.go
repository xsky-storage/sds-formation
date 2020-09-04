package utils

import (
	openapi_client "xsky.com/sds-formation/openapi-client"
)

// Define action types of resource
const (
	ActionTypeGet    = "Get"
	ActionTypeCreate = "Create"
	ActionTypeUpdate = "Update"
	ActionTypeDelete = "Delete"
)

// ResourceInterface resource interface
type ResourceInterface interface {
	Init(stack StackInterface)
	Repr() (repr interface{})
	GetType() (typeName string)
	CheckInterval() (interval int)

	IsReady() (ready bool)
	Get() (err error)
	Create() (created bool, err error)
	IsCreated() (created bool, err error)
	Update(repr interface{}) (updated bool, err error)
	IsUpdated() (updated bool, err error)
	Delete(repr interface{}) (deleted bool, err error)
	IsDeleted() (deleted bool, err error)
}

// StackInterface stack interface
type StackInterface interface {
	CallAPI(string, interface{}, map[string]string, ...map[string]string) ([]byte, error)
	GetOpenAPIClient() openapi_client.Client
	GetResourceValue(string) interface{}
}
