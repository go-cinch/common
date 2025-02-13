package id

import (
	"context"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/pkg/errors"
	"github.com/sony/sonyflake"
)

type Sonyflake struct {
	ops   SonyflakeOptions
	sf    *sonyflake.Sonyflake
	Error error
}

// NewSonyflake can get a unique code by id(You need to ensure that id is unique)
func NewSonyflake(options ...func(*SonyflakeOptions)) *Sonyflake {
	ops := getSonyflakeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	sf := &Sonyflake{
		ops: *ops,
	}
	st := sonyflake.Settings{
		StartTime: ops.startTime,
	}
	if ops.machineID > 0 {
		st.MachineID = func() (uint16, error) {
			return ops.machineID, nil
		}
	}
	ins := sonyflake.NewSonyflake(st)
	if ins == nil {
		sf.Error = errors.Errorf("create snoyflake failed")
	}
	_, err := ins.NextID()
	if err != nil {
		sf.Error = errors.Errorf("invalid start time")
	}
	sf.sf = ins
	return sf
}

func (s *Sonyflake) ID(ctx context.Context) (id uint64) {
	if s.Error != nil {
		log.WithContext(ctx).WithError(s.Error).Warn(s.Error)
		return
	}
	var err error
	id, err = s.sf.NextID()
	if err == nil {
		return
	}

	sleep := 1
	for {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		id, err = s.sf.NextID()
		if err == nil {
			return
		}
		sleep *= 2
	}
}
