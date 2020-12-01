package ad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Response struct {
	CodeError
	Data interface{} `json:"data"`
}

type CodeError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (ce CodeError) Error() string {
	return fmt.Sprintf("code: %d message: %s", ce.Code, ce.Message)
}

type Params []interface{}

func HttpPost(uri string, params Params, response interface{}) error {
	values, err := jsonUrlValues(params)
	if err != nil {
		return err
	}
	resp, err := http.PostForm(uri, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if data != nil {
		rsp := &Response{Data: response}
		err = json.Unmarshal(data, rsp)
		if err != nil {
			return err
		}
		if rsp.Code != 0 {
			return rsp.CodeError
		}
		return nil
	}
	return nil
}

func HttpGet(uri string, params Params, response interface{}) error {
	values, err := jsonUrlValues(params)
	if err != nil {
		return err
	}
	Url, err := url.Parse(uri)
	if err != nil {
		return err
	}
	query := Url.Query()
	for s, i := range values {
		query[s] = i
	}
	Url.RawQuery = query.Encode()
	resp, err := http.Get(Url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if data != nil {
		rsp := &Response{Data: response}
		err = json.Unmarshal(data, rsp)
		if err != nil {
			return err
		}
		if rsp.Code != 0 {
			return rsp.CodeError
		}
		return nil
	}
	return nil
}

// 将对象 v 转换成 url.Values, 如果字段中包含对象数组或对象, 则将其转换为 json 字符串
func jsonUrlValues(vs []interface{}) (url.Values, error) {
	var (
		obj  map[string]interface{}
		data []byte
		err  error
	)
	for _, v := range vs {
		data, err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &obj)
		if err != nil {
			return nil, err
		}
	}
	for key, v := range obj {
		switch sub := v.(type) {
		case map[string]interface{}, []interface{}:
			data, _ = json.Marshal(sub)
			obj[key] = string(data)
		}
	}
	var values = url.Values{}
	for key, v := range obj {
		values[key] = []string{fmt.Sprint(v)}
	}
	return values, nil
}
