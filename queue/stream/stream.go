package stream

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-cinch/common/log"
	"github.com/redis/go-redis/v9"
)

type Stream struct {
	ops Options
}

func New(options ...func(*Options)) *Stream {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return &Stream{
		ops: *ops,
	}
}

func (s *Stream) Pub(ctx context.Context, msg interface{}) error {
	rds := s.ops.rds
	key := s.ops.key
	m := make(map[string]interface{})
	str, err := json.Marshal(msg)
	if err != nil {
		log.WithContext(ctx).Warn("json.Marshal err: %v", err)
		return err
	}
	err = json.Unmarshal(str, &m)
	if err != nil {
		log.WithContext(ctx).Warn("json.Unmarshal err: %v", err)
		return err
	}
	pipe := rds.TxPipeline()
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: m,
	})
	if s.ops.expire > 0 {
		pipe.Expire(ctx, key, s.ops.expire)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.WithContext(ctx).Debug("pub err: %v", err)
	}
	return err
}

func (s *Stream) Read(ctx context.Context) <-chan map[string]interface{} {
	msgCh := make(chan map[string]interface{})
	rds := s.ops.rds
	key := s.ops.key
	go func() {
		defer close(msgCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msgs, err := rds.XRead(ctx, &redis.XReadArgs{
					Streams: []string{key, "0"},
					Count:   1,
					Block:   0, // 0 means block until have msg
				}).Result()
				if err != nil {
					log.WithContext(ctx).Debug("read err %s: %v", key, err)
					return
				}
				if len(msgs) != 1 || len(msgs[0].Messages) != 1 {
					log.WithContext(ctx).Warn("read err, invalid msg")
					return
				}
				select {
				case <-ctx.Done():
					return
				case msgCh <- msgs[0].Messages[0].Values:
				}
			}
		}
	}()
	return msgCh
}

func (s *Stream) ReadBatch(ctx context.Context, count int64) ([]redis.XMessage, error) {
	rds := s.ops.rds
	key := s.ops.key
	res, err := rds.XRead(ctx, &redis.XReadArgs{
		Streams: []string{key, "0"},
		Count:   count,
		Block:   20 * time.Second,
	}).Result()
	msgs := make([]redis.XMessage, 0, count)
	for _, v1 := range res {
		for _, v2 := range v1.Messages {
			msgs = append(msgs, v2)
		}
	}
	return msgs, err
}

func (s *Stream) Sub(ctx context.Context, done, cancel <-chan struct{}) <-chan redis.XMessage {
	msgCh := make(chan redis.XMessage)
	rds := s.ops.rds
	key := s.ops.key
	group := s.ops.group
	consumer := s.ops.consumer
	go func() {
		defer close(msgCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-cancel:
				return
			default:
				res, err := rds.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    group,
					Consumer: consumer,
					Streams:  []string{key, ">"},
					Count:    1,
					Block:    0,
				}).Result()
				if err != nil {
					log.WithContext(ctx).Debug("XReadGroup err %s: %v", key, err)
					time.Sleep(10 * time.Millisecond)
					continue
				}
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				case <-cancel:
					return
				case msgCh <- res[0].Messages[0]:
				}
			}
		}
	}()
	return msgCh
}

func (s *Stream) Ack(ctx context.Context, ids ...string) {
	rds := s.ops.rds
	stream := s.ops.key
	group := s.ops.group
	rds.XAck(ctx, stream, group, ids...)
}

func (s *Stream) Trim(ctx context.Context, max int64) {
	rds := s.ops.rds
	stream := s.ops.key
	rds.XTrimMaxLen(ctx, stream, max)
}

func (s *Stream) TrimLteMinID(ctx context.Context, id string) {
	rds := s.ops.rds
	stream := s.ops.key
	rds.XTrimMinID(ctx, stream, id)
	rds.XDel(ctx, stream, id)
}
