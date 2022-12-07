package rabbit

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
	"time"
)

func TestExchange_PublishProto(t *testing.T) {
	rb := New(
		uri,
	)
	if rb.Error != nil {
		panic(rb.Error)
	}
	ex := rb.Exchange(
		WithExchangeName("ex1"),
		WithExchangeDeclare(false),
	)
	if ex.Error != nil {
		panic(ex.Error)
	}

	for {
		time.Sleep(100 * time.Millisecond)
		go func() {
			var mqPb emptypb.Empty
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt1"),
				WithPublishRouteKey("rt2"),
				WithPublishCtx(context.Background()),
			)
			fmt.Println(time.Now(), "send 1 end", err)
		}()
		go func() {
			var mqPb emptypb.Empty
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt2"),
				WithPublishCtx(context.Background()),
			)
			fmt.Println(time.Now(), "send 2 end", err)
		}()
		go func() {
			var mqPb emptypb.Empty
			err := ex.PublishProto(
				&mqPb,
				WithPublishRouteKey("rt2"),
				WithPublishCtx(context.Background()),
			)
			fmt.Println(time.Now(), "send 3 end", err)
		}()
		fmt.Println()
	}

}

func TestExchange_PublishProto2(t *testing.T) {
	rb := New(
		uri,
	)
	if rb.Error != nil {
		panic(rb.Error)
	}
	ex := rb.Exchange(
		WithExchangeName("ex1"),
		WithExchangeDeclare(false),
	)
	if ex.Error != nil {
		panic(ex.Error)
	}

	for {
		time.Sleep(time.Second)
		var mqPb emptypb.Empty
		err := ex.PublishProto(
			&mqPb,
			WithPublishRouteKey("rt1"),
		)
		fmt.Println(time.Now(), "send end", err)
	}
}
