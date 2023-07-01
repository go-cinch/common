package minio

import "github.com/pkg/errors"

var (
	ErrEndpointNil        = errors.New("endpoint is empty")
	ErrKeyNil             = errors.New("key is empty")
	ErrSecretNil          = errors.New("secret is empty")
	ErrInitializeFailed   = errors.New("initialize failed")
	ErrObjectNameNil      = errors.New("object name is empty")
	ErrExpireInvalid      = errors.New("token expire time invalid")
	ErrContentTypeInvalid = errors.New("object content type invalid")
	ErrObjectSizeInvalid  = errors.New("object size invalid")
	ErrTokenSignFailed    = errors.New("sign token failed")
)
