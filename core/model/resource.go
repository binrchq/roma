package model

import "bitrec.ai/roma/core/types"

type Resource interface {
	GetResource() Resource
	GetConnect() []*types.Connection
	GetID() int64
	GetName() string
	GetLine() []string
	GetTitle() []string
}
