package utils

import (
	"net/http"

	"github.com/pkg/errors"
)

func HttpGet(client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "创建GET请求失败！")
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		// 当这里的 err 不为空时，res 可能会直接为 nil，因此就不能在这里 close
		return nil, errors.Wrap(err, "发送GET请求失败！")
	}

	if res.StatusCode != http.StatusOK {
		res.Body.Close() // 确保在非 200 OK 响应时关闭资源
		return nil, errors.Wrap(errors.Errorf("网络请求失败，错误码为：%d", res.StatusCode), "HTTP状态码不为200")
	}

	return res, nil
}
