package env

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/songs/src/platform"
	"github.com/morgine/songs/src/wpt"
	"time"
)

type accessTokenStorage struct {
	client    *redis.Client
	keyPrefix string
}

func newAccessTokenStorage(keyPrefix string, store *redis.Client) platform.AccessTokenStorage {
	return &accessTokenStorage{client: store, keyPrefix: keyPrefix}
}

func (r *accessTokenStorage) key(key string) string {
	return r.keyPrefix + key
}

func (r *accessTokenStorage) SetAccessToken(key string, token *platform.AccessToken) error {
	now := platform.Now().Unix()
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), r.key(key), string(data), time.Duration(token.RefreshExpireAt-now)*time.Second).Err()
}

func (r *accessTokenStorage) GetAccessToken(key string) (token *platform.AccessToken, err error) {
	data, err := r.client.Get(context.Background(), r.key(key)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		token = &platform.AccessToken{}
		err = json.Unmarshal(data, token)
		if err != nil {
			return nil, err
		} else {
			return token, nil
		}
	}
}

type openPlatformStorage struct {
	platform.AccessTokenStorage
	gorm *model.AppModel
}

func newOpenPlatformStorage(tokenStore platform.AccessTokenStorage, gorm *model.AppModel) *openPlatformStorage {
	return &openPlatformStorage{
		AccessTokenStorage: tokenStore,
		gorm:               gorm,
	}
}

func (r *openPlatformStorage) SaveAuthorizer(authorizer *wpt.Authorizer) error {
	app := authorizerToApp(authorizer)[0]
	return r.gorm.SaveAPP(app.Appid, app)
}

// 重置所有数据库中的 app 信息为当提供的数据
func (r *openPlatformStorage) ResetAuthorizers(authroizers []*wpt.Authorizer) error {
	return r.gorm.ResetApps(authorizerToApp(authroizers...))
}

func (r *openPlatformStorage) DeleteAuthorizer(appid string) error {
	return r.gorm.DelAppByAppid(appid)
}

func (r *openPlatformStorage) GetAuthorizersByAppids(appids []string) ([]*model.App, error) {
	return r.gorm.GetAppsByAppids(appids)
}

func (r *openPlatformStorage) GetAuthorizers() ([]*model.App, error) {
	return r.gorm.GetApps()
}

func authorizerToApp(authorizer ...*wpt.Authorizer) []*model.App {
	var apps = make([]*model.App, len(authorizer))
	for i, a := range authorizer {
		info := a.AuthorizerInfo
		apps[i] = &model.App{
			ID:            0,
			Appid:         a.AuthorizationInfo.AuthorizerAppid,
			NickName:      info.NickName,
			HeadImg:       info.HeadImg,
			UserName:      info.UserName,
			PrincipalName: info.PrincipalName,
			Alias:         info.Alias,
			QrcodeUrl:     info.QrcodeUrl,
			Signature:     info.Signature,
		}
	}
	return apps
}
