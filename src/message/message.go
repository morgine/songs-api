package message

import (
	"fmt"
)

const StatusOK Status = 200

const (
	ErrUnknown Status = 404 + iota
	ErrAdminUsernameAlreadyExist
	ErrAdminUsernameOrPasswordIncorrect
	ErrAdminUnauthorized
	ErrAdvertPlatformUnauthorized
)

var StatusText = map[Status]string{
	StatusOK:                            "操作成功",
	ErrUnknown:                          "其他错误",
	ErrAdminUsernameAlreadyExist:        "管理员账户已存在",
	ErrAdminUsernameOrPasswordIncorrect: "管理账户或密码错误",
	ErrAdminUnauthorized:                "需要管理员授权",
	ErrAdvertPlatformUnauthorized:       "广告平台未授权",
}

type Status int

func (c Status) Error() string {
	return fmt.Sprintf("error code %d: %s", c, StatusText[c])
}
