package env

import (
	"github.com/morgine/pkg/config"
	"github.com/morgine/songs/src/platform"
)

func NewAdvertPlatform(namespace string, accessStore platform.AccessTokenStorage, configs config.Configs) (*platform.AdvertPlatform, error) {
	env := &AdvertPlatformEnv{}
	err := configs.UnmarshalSub(namespace, env)
	if err != nil {
		return nil, err
	}
	return platform.NewAdvertPlatform(env.ClientID, env.Secret, accessStore), nil
}
