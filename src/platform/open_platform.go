package platform

import (
	"errors"
	"fmt"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/songs/src/wpt"
	"github.com/morgine/songs/src/wpt/message"
	"net/http"
	"sync"
	"time"
)

var tenYearSeconds int64 = 10 * 365 * 24 * 3600 // 10 年

type OpenPlatform struct {
	appid          string
	appSecret      string
	msgVerifyToken string
	decrypter      *wpt.WXBizMsgCrypt
	notify         *wpt.AuthorizationNotify
	store          OpenPlatformStorage
	mu             sync.Mutex
}

func NewOpenPlatform(appid, secret string, store OpenPlatformStorage, msgVerifyToken string, decrypter *wpt.WXBizMsgCrypt) *OpenPlatform {
	return &OpenPlatform{
		appid:          appid,
		appSecret:      secret,
		decrypter:      decrypter,
		msgVerifyToken: msgVerifyToken,
		notify:         nil,
		store:          store,
		mu:             sync.Mutex{},
	}
}

// 读取用户发送/触发的消息, 如果 decrypter 不为 nil, 则通过 decrypter 解密, 否则按明文方式解析消息.
func (w *OpenPlatform) ListenMessage(r *http.Request) (msg *message.ServerMessage, echoStr string, err error) {
	echoStr, err = message.CheckSignature(r, w.msgVerifyToken)
	if err != nil {
		return nil, "", err
	}
	if echoStr != "" {
		return nil, echoStr, nil
	} else {
		msg, err = message.ReadServerMessage(r, w.decrypter)
	}
	return
}

// Response 用于被动回复消息, 当用户发送文本、图片、视频、图文、地理位置这五种消息时，开发者只能回复1条
// 图文消息；其余场景最多可回复8条图文消息, 多余的消息将被忽略
func (w *OpenPlatform) ResponseMessage(serverMsg *message.ServerMessage, respMsg *message.ResponseMessage, writer http.ResponseWriter) error {
	return message.Response(serverMsg, respMsg, writer, w.decrypter)
}

// 监听第三方授权通知, 处理完毕后返回 success, 否则微信会一直重复发送该消息
func (w *OpenPlatform) ListenTicket(r *http.Request) error {
	notify, err := wpt.ListenComponentAuthorizationNotify(r, w.decrypter)
	if err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.notify = notify
	switch notify.InfoType {
	// 取消授权
	case wpt.EvtUnauthorized:
		err = w.store.DeleteAuthorizer(notify.AuthorizerAppid)
		if err != nil {
			return err
		}
	// 更新授权
	case wpt.EvtUpdateauthorized, wpt.EvtAuthorized:
		// 获取并保存授权方信息
		//var authorizer *wpt.Authorizer
		//authorizer, err = w.getAuthorizerInfo(notify.AuthorizerAppid)
		//if err != nil {
		//	return err
		//}
		//err = w.store.SaveAuthorizer(authorizer)
		//if err != nil {
		//	return err
		//}
		//now := Now().Unix()
		//authInfo := authorizer.AuthorizationInfo.AuthorizerToken
		//err = w.store.SetAccessToken(notify.AuthorizerAppid, &AccessToken{
		//	Token:           authInfo.AuthorizerAccessToken,
		//	Refresh:         authInfo.AuthorizerRefreshToken,
		//	TokenExpireAt:   now + (authInfo.ExpiresIn - (authInfo.ExpiresIn >> 3)), // 提前 1/8 过期时间
		//	RefreshExpireAt: now + tenYearSeconds,
		//})
		//if err != nil {
		//	return err
		//}
	}
	return nil
}

// 需要上锁，需要用到 veryTicket
func (w *OpenPlatform) GetComponentAccessToken() (string, error) {
	token, err := w.store.GetAccessToken(w.appid + "_cat")
	if err != nil {
		return "", err
	}
	now := Now().Unix()
	if token == nil || token.TokenExpireAt < now {
		var at *wpt.ComponentAccessToken
		at, err = wpt.GetComponentAccessToken(w.appid, w.appSecret, w.notify.ComponentVerifyTicket)
		if err != nil {
			return "", err
		}
		token = &AccessToken{
			Token:           at.Token,
			Refresh:         "",
			TokenExpireAt:   now + (at.ExpiresIn - (at.ExpiresIn >> 3)), // 过期时间提前 1/8
			RefreshExpireAt: 0,
		}
		token.RefreshExpireAt = token.TokenExpireAt // 数据保留时间，无 refresh token, 因此按 token 最长时间保留数据
		err = w.store.SetAccessToken(w.appid+"_cat", token)
		if err != nil {
			return "", err
		}
	}
	return token.Token, nil
}

// 需要用到 access token
func (w *OpenPlatform) GetPreAuthCode() (string, error) {
	code, err := w.store.GetAccessToken(w.appid + "_pac")
	if err != nil {
		return "", err
	}
	now := Now().Unix()
	if code == nil || code.TokenExpireAt < now {
		var token string
		token, err = w.GetComponentAccessToken()
		if err != nil {
			return "", err
		}
		var preAuthCode *wpt.PreAuthCode
		preAuthCode, err = wpt.CreatePreAuthCode(w.appid, token)
		if err != nil {
			return "", err
		}
		code = &AccessToken{
			Token:           preAuthCode.PreAuthCode,
			Refresh:         "",
			TokenExpireAt:   now + (preAuthCode.ExpiresIn - (preAuthCode.ExpiresIn >> 3)),
			RefreshExpireAt: 0,
		}
		code.RefreshExpireAt = code.TokenExpireAt // 数据保留时间，无 refresh token, 因此按 token 最长时间保留数据
		err = w.store.SetAccessToken(w.appid+"_pac", code)
		if err != nil {
			return "", err
		}
	}
	return code.Token, nil
}

func (w *OpenPlatform) OpenPlatformAuthRedirectUrl(redirectUrl string) (string, error) {
	preAuthCode, err := w.GetPreAuthCode()
	if err != nil {
		return "", err
	}
	opt := &wpt.AuthOption{
		ComponentAppid: w.appid,
		PreAuthCode:    preAuthCode,
		RedirectUri:    redirectUrl,
		AuthType:       "1",
		BizAppid:       "",
	}
	return wpt.NewAuthRedirectUrl(opt), nil
}

func (w *OpenPlatform) getAuthCode(r *http.Request) (string, error) {
	code, _ := wpt.GetAuthCode(r)
	if code == "" {
		return "", errors.New("获取授权码失败")
	}
	return code, nil
}

func (w *OpenPlatform) getAuthorizationInfo(authCode string) (*wpt.AuthorizationInfo, error) {
	accessToken, err := w.GetComponentAccessToken()
	if err != nil {
		return nil, err
	}
	return wpt.GetAuthorizationInfo(w.appid, authCode, accessToken)
}

func (w *OpenPlatform) ListenAuthorized(r *http.Request) error {
	authCode, err := w.getAuthCode(r)
	if err != nil {
		return err
	}
	authInfo, err := w.getAuthorizationInfo(authCode)
	if err != nil {
		return err
	}
	err = w.store.SetAccessToken(authInfo.AuthorizerAppid, &AccessToken{
		Token:           authInfo.AuthorizerAccessToken,
		Refresh:         authInfo.AuthorizerRefreshToken,
		TokenExpireAt:   Now().Add(time.Duration(authInfo.ExpiresIn - (authInfo.ExpiresIn >> 3))).Unix(),
		RefreshExpireAt: tenYearSeconds,
	})
	if err != nil {
		return err
	}

	// 测试 access token 是否能被获取
	_, err = w.GetAuthorizerAccessToken(authInfo.AuthorizerAppid)
	if err != nil {
		return err
	}

	// 获取并保存授权方信息
	authorizer, err := w.getAuthorizerInfo(authInfo.AuthorizerAppid)
	if err != nil {
		return err
	}
	return w.store.SaveAuthorizer(authorizer)
}

func (w *OpenPlatform) GetAuthorizerAccessToken(appid string) (accessToken string, err error) {
	token, err := w.store.GetAccessToken(appid)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", fmt.Errorf("公众号未授权或授权信息丢失，需要重新授权")
	}
	now := Now().Unix()
	if token.TokenExpireAt < now {
		accessToken, err := w.GetComponentAccessToken()
		if err != nil {
			return "", err
		}
		authToken, err := wpt.RefreshAuthorizerToken(w.appid, appid, token.Refresh, accessToken)
		if err != nil {
			return "", err
		}
		token = &AccessToken{
			Token:           authToken.AuthorizerAccessToken,
			Refresh:         authToken.AuthorizerRefreshToken,
			TokenExpireAt:   Now().Add(time.Duration(authToken.ExpiresIn - (authToken.ExpiresIn >> 3))).Unix(),
			RefreshExpireAt: tenYearSeconds,
		}
		err = w.store.SetAccessToken(appid, token)
		if err != nil {
			return "", err
		}
	}
	return token.Token, nil
}

func (w *OpenPlatform) getAuthorizerInfo(appid string) (*wpt.Authorizer, error) {
	accessToken, err := w.GetComponentAccessToken()
	if err != nil {
		return nil, err
	}
	return wpt.GetAuthorizerInfo(w.appid, appid, accessToken)
}

// 重置已授权公众号列表
func (w *OpenPlatform) ResetAuthorizers() error {
	accessToken, err := w.GetComponentAccessToken()
	if err != nil {
		return err
	}
	// 包含所有 app 信息，一次性全部重置
	var authorizers []*wpt.Authorizer
	var offset, limit = 0, 100

	for {
		list, err := wpt.GetAuthorizerList(accessToken, w.appid, offset, limit)
		if err != nil {
			return err
		}
		if offset >= list.TotalCount {
			break
		} else {
			offset += limit
		}
		for _, information := range list.List {
			token := &AccessToken{
				Token:           "",
				Refresh:         information.RefreshToken,
				TokenExpireAt:   0,
				RefreshExpireAt: tenYearSeconds,
			}
			// 保存 access token
			err = w.store.SetAccessToken(information.AuthorizerAppid, token)
			if err != nil {
				return err
			}
			// 获取授权方信息
			authorizer, err := w.getAuthorizerInfo(information.AuthorizerAppid)
			if err != nil {
				return err
			}
			authorizers = append(authorizers, authorizer)
		}
	}
	return w.store.ResetAuthorizers(authorizers)
}

type Summary struct {
	Total        []*wpt.UserSummary
	AppSummaries []*AppSummary
}

type AppSummary struct {
	Appid     string
	Nickname  string
	Err       string
	Summaries []*wpt.UserSummary
}

func (w *OpenPlatform) GetAppSummaries(appids []string, beginDate, endDate string) (summaries []*AppSummary, err error) {
	var apps []*model.App
	if len(appids) == 0 {
		apps, err = w.store.GetAuthorizers()
		if err != nil {
			return nil, err
		}
	} else {
		apps, err = w.store.GetAuthorizersByAppids(appids)
		if err != nil {
			return nil, err
		}
	}
	for _, app := range apps {
		appid := app.Appid
		var summary = &AppSummary{
			Appid:    appid,
			Nickname: app.NickName,
		}
		summaries = append(summaries, summary)
		accessToken, err := w.GetAuthorizerAccessToken(appid)
		if err != nil {
			summary.Err = err.Error()
		} else {
			var userSummaries []*wpt.UserSummary
			userSummaries, err = wpt.GetUserSummary(accessToken, beginDate, endDate)
			if err != nil {
				summary.Err = err.Error()
			} else {
				summary.Summaries = userSummaries
			}
		}
	}
	return summaries, nil
}

// 获取用户增减数据, 该操作很费时间, 且接口调用次数有限
func (w *OpenPlatform) GetUserSummary(appids []string, beginDate, endDate string) (summary *Summary, err error) {
	appSummaries, err := w.GetAppSummaries(appids, beginDate, endDate)
	if err != nil {
		return nil, err
	}
	return CountTotalSummary(appSummaries), nil
}

func CountTotalSummary(appSummaries []*AppSummary) *Summary {
	var totalSummaries []*wpt.UserSummary
	// 累加
	for _, appSummary := range appSummaries {
		for _, userSummary := range appSummary.Summaries {
			func() {
				for _, totalSummary := range totalSummaries {
					if totalSummary.RefDate == userSummary.RefDate {
						totalSummary.CancelUser += userSummary.CancelUser
						totalSummary.CumulateUser += userSummary.CumulateUser
						totalSummary.NewUser += userSummary.NewUser
						return
					}
				}
				totalSummaries = append(totalSummaries, &wpt.UserSummary{
					RefDate:      userSummary.RefDate,
					UserSource:   0,
					NewUser:      userSummary.NewUser,
					CancelUser:   userSummary.CancelUser,
					CumulateUser: userSummary.CumulateUser,
				})
			}()
		}
	}
	return &Summary{
		Total:        totalSummaries,
		AppSummaries: appSummaries,
	}
}

type Cumulate struct {
	Total       []*wpt.UserCumulate
	AppCumulate []*AppCumulate
}

type AppCumulate struct {
	Appid     string
	Nickname  string
	Cumulates []*wpt.UserCumulate
}

// 获取累计用户数据
func (w *OpenPlatform) GetUserCumulate(beginDate, endDate string) (cumulate *Cumulate, err error) {
	apps, err := w.store.GetAuthorizers()
	if err != nil {
		return nil, err
	}
	cumulate = &Cumulate{}
	for _, app := range apps {
		appid := app.Appid
		accessToken, err := w.GetAuthorizerAccessToken(appid)
		if err != nil {
			return nil, err
		}
		var cumulates []*wpt.UserCumulate
		cumulates, err = wpt.GetUserCumulate(accessToken, beginDate, endDate)
		if err != nil {
			return nil, err
		}
		if len(cumulates) > 0 {
			nickName := app.NickName
			cumulate.AppCumulate = append(cumulate.AppCumulate, &AppCumulate{
				Appid:     appid,
				Nickname:  nickName,
				Cumulates: cumulates,
			})
			// 累加
			for _, userCumulate := range cumulates {
				func() {
					for _, userSummary := range cumulate.Total {
						if userSummary.RefDate == userSummary.RefDate {
							userSummary.CumulateUser += userSummary.CumulateUser
							return
						}
					}
					cumulate.Total = append(cumulate.Total, userCumulate)
				}()
			}
		}
	}
	return cumulate, nil
}

// 创建公众号标签
func (w *OpenPlatform) CreateAppUserTag(appid string, tagName string) (*wpt.UserTag, error) {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return nil, err
	}
	return wpt.CreateAppUserTag(accessToken, tagName)
}

// 获得公众号标签
func (w *OpenPlatform) GetAppUserTags(appid string) ([]*wpt.UserTag, error) {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return nil, err
	}
	return wpt.GetAppUserTags(accessToken)
}

// 更新公众号标签
func (w *OpenPlatform) UpdateAppUserTag(appid string, tag *wpt.UserTag) error {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return err
	}
	return wpt.UpdateAppUserTag(accessToken, tag)
}

// 删除公众号标签
func (w *OpenPlatform) DeleteAppUserTag(appid string, tagID int) error {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return err
	}
	return wpt.DeleteAppUserTag(accessToken, tagID)
}

// 获得对应标签下的用户列表
func (w *OpenPlatform) GetAppTagUsers(appid string, tagID int, nextOpenid string) (*wpt.TagUsers, error) {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return nil, err
	}
	return wpt.GetAppTagUsers(accessToken, tagID, nextOpenid)
}

// 批量为用户打标签
func (w *OpenPlatform) BatchTagging(appid string, tagID int, openids []string) error {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return err
	}
	return wpt.BatchTagging(accessToken, tagID, openids)
}

// 批量为用户取消标签
func (w *OpenPlatform) BatchUntagging(appid string, tagID int, openids []string) error {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return err
	}
	return wpt.BatchUntagging(accessToken, tagID, openids)
}

// 获取用户身上的标签列表
func (w *OpenPlatform) GetUserTags(appid string, openid string) (ids []int, err error) {
	accessToken, err := w.GetAuthorizerAccessToken(appid)
	if err != nil {
		return nil, err
	}
	return wpt.GetUserTags(accessToken, openid)
}
