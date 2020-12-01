package platform

import (
	"github.com/morgine/songs/src/model"
	"github.com/morgine/songs/src/wpt"
	"time"
)

var Now = time.Now

type AccessToken struct {
	Token           string
	Refresh         string
	TokenExpireAt   int64
	RefreshExpireAt int64
}

type AccessTokenStorage interface {
	SetAccessToken(key string, token *AccessToken) error
	GetAccessToken(key string) (token *AccessToken, err error)
}

type OpenPlatformStorage interface {
	AccessTokenStorage
	SaveAuthorizer(authorizer *wpt.Authorizer) error
	ResetAuthorizers(authroizers []*wpt.Authorizer) error
	DeleteAuthorizer(appid string) error
	GetAuthorizers() ([]*model.App, error)
	GetAuthorizersByAppids(appids []string) ([]*model.App, error)
}

//type Storage interface {
//	SetValue(key string, v interface{}, expires time.Duration) error
//	GetValue(key string, v interface{}) error
//}
//
//type RedisStorage struct {
//	client *redis.Client
//}
//
//func NewRedisStorage(client *redis.Client) *RedisStorage {
//	return &RedisStorage{client: client}
//}
//
//func (r *RedisStorage) SetValue(key string, v interface{}, expires time.Duration) error {
//	data, err := json.Marshal(v)
//	if err != nil {
//		return err
//	}
//	return r.client.Set(context.Background(), key, string(data), expires).Err()
//}
//
//func (r *RedisStorage) GetValue(key string, v interface{}) error {
//	data, err := r.client.Get(context.Background(), key).Bytes()
//	if err != nil {
//		return err
//	}
//	return json.Unmarshal(data, v)
//}

//type accessTokenStorage struct {
//	store Storage
//}
//
//func NewAccessTokenStorage(sotre Storage) *accessTokenStorage {
//	return &accessTokenStorage{store: sotre}
//}
//
//func (r *accessTokenStorage) SetAccessToken(key string, token *AccessToken) error {
//	now := Now().Unix()
//	return r.store.SetValue(key, token, time.Duration(token.RefreshExpireAt-now)*time.Second)
//}
//
//func (r *accessTokenStorage) GetAccessToken(key string) (token *AccessToken, err error) {
//	token = &AccessToken{}
//	err = r.store.GetValue(key, token)
//	return
//}
//
//type OpenPlatformStore struct {
//	accessTokenStorage
//}
//
//
//func (r *OpenPlatformStore) SaveAuthorizer(authorizer *wpt.Authorizer) error {
//	return model.SaveApp(authorizerToApp(authorizer)[0])
//}
//
//func (r *OpenPlatformStore) ResetAuthorizers(authroizers []*wpt.Authorizer) error {
//	return model.ResetApps(authorizerToApp(authroizers...))
//}
//
//func (r *OpenPlatformStore) DeleteAuthorizer(appid string) error {
//	return model.DeleteApp(appid)
//}
//
//func (r *OpenPlatformStore) GetAuthorizers() ([]*model.App, error) {
//	return model.GetAllApps()
//}
//
//func authorizerToApp(authorizer ...*wpt.Authorizer) []*model.App {
//	var apps = make([]*model.App, len(authorizer))
//	for i, a := range authorizer {
//		info := a.AuthorizerInfo
//		apps[i] = &model.App{
//			ID:            0,
//			Appid:         a.AuthorizationInfo.AuthorizerAppid,
//			NickName:      info.NickName,
//			HeadImg:       info.HeadImg,
//			UserName:      info.UserName,
//			PrincipalName: info.PrincipalName,
//			Alias:         info.Alias,
//			QrcodeUrl:     info.QrcodeUrl,
//			Signature:     info.Signature,
//		}
//	}
//	return apps
//}