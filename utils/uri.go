package utils

import (
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"net/url"
	"strconv"
	"strings"
)

func ParseRedisURI(uri string) (client redis.UniversalClient, err error) {
	if uri != "" {
		var u *url.URL
		u, err = url.Parse(uri)
		if err != nil {
			err = errors.Errorf("invalid redis uri %s: %v", uri, err)
			return
		}
		q := u.Query()

		addrs := strings.Split(u.Host, ",")
		var master, username, password, sentinelUsername, sentinelPassword string
		var db int
		master = q.Get("master")
		username = u.User.Username()
		if v, ok := u.User.Password(); ok {
			password = v
		}

		if len(u.Path) > 0 {
			xs := strings.Split(strings.Trim(u.Path, "/"), "/")
			db, err = strconv.Atoi(xs[0])
			if err != nil {
				err = errors.Errorf("invalid db %s: %v", uri, err)
				return
			}
		}

		sentinel, _ := strconv.ParseBool(q.Get("sentinel"))
		if sentinel {
			sentinelUsername = username
			sentinelPassword = password
			username = ""
			password = ""
		}
		client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:            addrs,
			MasterName:       master,
			DB:               db,
			Username:         username,
			Password:         password,
			SentinelUsername: sentinelUsername,
			SentinelPassword: sentinelPassword,
		})
		return
	}
	err = errors.Errorf("invalid redis uri")
	return
}
