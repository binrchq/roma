package model

type Resource interface {
	GetResource() Resource
	GetConnect() []map[string]interface{}
	GetID() int64
	GetName() string
	GetLine() []string
	GetTitle() []string
}
