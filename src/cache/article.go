package cache

import "encoding/json"

type Article struct {
	Title       string
	Description string
	Url         string
	PicFile     string
}

type ArticleClient struct {
	keyPrefix string
	engine    Engine
}

func NewArticleClient(keyPrefix string, engine Engine) *ArticleClient {
	return &ArticleClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *ArticleClient) Get(componentAppid string) ([]*Article, error) {
	data, err := tm.engine.Get(tm.key(componentAppid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		var articles []*Article
		err = json.Unmarshal(data, &articles)
		if err != nil {
			return nil, err
		} else {
			return articles, nil
		}
	} else {
		return nil, nil
	}
}

func (tm *ArticleClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *ArticleClient) Set(componentAppid string, articles []*Article) error {
	data, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(componentAppid), data, 0)
}

func (tm *ArticleClient) Del(componentAppid string) error {
	return tm.engine.Del(tm.key(componentAppid))
}
