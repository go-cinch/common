package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/alicebob/miniredis/v2"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
)

func NewMySQL() (string, int, error) {
	host := "0.0.0.0"
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := random.Intn(3000) + 3000
	engine := sqle.NewDefault(memory.NewDBProvider())
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", host, port),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		return "", 0, err
	}

	go func() {
		err = s.Start()
		if err != nil {
			return
		}
	}()
	return host, port, err
}

func NewRedis() (string, int, error) {
	host := "0.0.0.0"
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := random.Intn(3000) + 6000
	s := miniredis.NewMiniRedis()
	err := s.StartAddr(fmt.Sprintf("%s:%d", host, port))
	return host, port, err
}
