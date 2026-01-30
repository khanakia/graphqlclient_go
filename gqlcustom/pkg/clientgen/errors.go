package clientgen

import "errors"

var (
	ErrSchemaPathRequired = errors.New("schema path is required")
	ErrConfigPathRequired = errors.New("config path is required")
	ErrSchemaNotFound     = errors.New("schema file not found")
	ErrSchemaParseFailed  = errors.New("failed to parse schema")
)
