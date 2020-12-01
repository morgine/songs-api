package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

type AdminSession struct {
	client    *redis.Client
	keyPrefix string
}

func NewAdminSession(keyPrefix string, client *redis.Client) *AdminSession {
	return &AdminSession{client: client, keyPrefix: keyPrefix}
}

func (a *AdminSession) key(adminID int) string {
	return a.keyPrefix + strconv.Itoa(adminID)
}

func (a *AdminSession) SaveAuthToken(adminID int, token string, expires int64) error {
	return a.client.Set(context.Background(), a.key(adminID), token, time.Duration(expires)*time.Second).Err()
}

func (a *AdminSession) CheckAndRefreshToken(adminID int, token string, expires int64) (ok bool, err error) {
	key := a.key(adminID)
	savedToken, err := a.client.Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if savedToken != token {
		return false, nil
	} else {
		err = a.client.Expire(context.Background(), key, time.Duration(expires)*time.Second).Err()
		if err != nil {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (a *AdminSession) DelAuthToken(adminID int) error {
	return a.client.Del(context.Background(), a.key(adminID)).Err()
}
