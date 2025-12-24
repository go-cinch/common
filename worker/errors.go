package worker

import "fmt"

var (
	ErrUUIDNil                       = fmt.Errorf("uuid is empty")
	ErrRedisNil                      = fmt.Errorf("redis is empty")
	ErrRedisInvalid                  = fmt.Errorf("redis is invalid")
	ErrExprInvalid                   = fmt.Errorf("expr is invalid")
	ErrSaveCron                      = fmt.Errorf("save cron failed")
	ErrHTTPCallbackInvalidStatusCode = fmt.Errorf("http callback invalid status code")
	ErrCronTaskNotFound              = fmt.Errorf("cron task not found")
)
