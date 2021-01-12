package ad

import (
	"github.com/google/go-querystring/query"
	"net/http"
)

type AuthRedirectOptions struct {
	// 开发者创建的应用程序的唯一标识id，必填，可通过【应用程序管理页面】查看
	ClientID string `url:"client_id,omitempty"`
	// 回调地址，由开发者自行提供和定义（地址主域需与开发者创建应用程序时登记的回调主域一致），
	// 用于跳转并接收回调信息，必填，该字段需要做UrlEncode，确保后续跳转正常；
	RedirectUri string `url:"redirect_uri,omitempty"`
	// 开发者自定义参数，可用于验证请求有效性或者传递其他自定义信息，回调的时候会原样带回，可选；
	State string `url:"state,omitempty"`
	// 授权的能力范围，可选，不传时表示授权范围为当前应用程序所拥有的所有全部权限；
	Scope string `url:"scope,omitempty"`
	// 授权用的账号类型，可选，包括QQ号和微信号，不传时默认为QQ号
	AccountType string `url:"account_type,omitempty"`
}

func AuthRedirect(opts *AuthRedirectOptions) string {
	vs, _ := query.Values(opts)
	return "https://developers.e.qq.com/oauth/authorize?" + vs.Encode()
}

// 用户授权后跳转至指定的 URI 时会带上, code 及 state 参数, 如果用户未同一授权, 则只会获得 state 参数
func GetAuthorizationCode(req *http.Request) (authorizationCode, state string) {
	q := req.URL.Query()
	return q.Get("authorization_code"), q.Get("state")
}

type AccessToken struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

type GetAccessTokenOptions struct {
	// 应用 id，在开发者官网创建应用后获得，可通过 [应用程序管理页面] 查看
	ClientID string `url:"client_id,omitempty"`
	// 应用 secret，在开发者官网创建应用后获得，可通过 [应用程序管理页面] 查看
	// 字段长度最小 1 字节，长度最大 256 字节
	ClientSecret string `url:"client_secret,omitempty"`
	// 请求的类型，可选值： authorization_code （授权码方式获取 token ）、 refresh_token （刷新 token ）
	// 字段长度最小 1 字节，长度最大 64 字节
	AuthorizationCode string `url:"authorization_code,omitempty"`
	// 应用回调地址，当 grant_type=authorization_code 时， redirect_uri 为必传参数，仅支持 http 和 https，
	// 不支持指定端口号，且传入的地址需要与获取 authorization_code 时，传入的回调地址保持一致
	// 字段长度最小 1 字节，长度最大 1024 字节
	RedirectUri string `url:"redirect_uri,omitempty"`
}

func (o *GetAccessTokenOptions) uri() (string, error) {
	vs, err := query.Values(o)
	if err != nil {
		return "", err
	}
	vs.Set("grant_type", "authorization_code")
	return "https://api.e.qq.com/oauth/token?" + vs.Encode(), nil
}

func GetAccessToken(opts *GetAccessTokenOptions) (*AccessToken, error) {
	uri, err := opts.uri()
	if err != nil {
		return nil, err
	}
	token := &AccessToken{}
	err = HttpGet(uri, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

type RefreshAccessTokenOptions struct {
	// 应用 id，在开发者官网创建应用后获得，可通过 [应用程序管理页面] 查看
	ClientID string `url:"client_id,omitempty"`
	// 应用 secret，在开发者官网创建应用后获得，可通过 [应用程序管理页面] 查看
	// 字段长度最小 1 字节，长度最大 256 字节
	ClientSecret string `url:"client_secret,omitempty"`
	// 应用 refresh token，当 grant_type=refresh_token 时必填
	// 字段长度最小 1 字节，长度最大 256 字节
	RefreshToken string `url:"refresh_token,omitempty"`
}

func (o *RefreshAccessTokenOptions) uri() (string, error) {
	vs, err := query.Values(o)
	if err != nil {
		return "", err
	}
	vs.Set("grant_type", "refresh_token")
	return "https://api.e.qq.com/oauth/token?" + vs.Encode(), nil
}

func RefreshAccessToken(opts *RefreshAccessTokenOptions) (*AccessToken, error) {
	uri, err := opts.uri()
	if err != nil {
		return nil, err
	}
	token := &AccessToken{}
	err = HttpGet(uri, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}
