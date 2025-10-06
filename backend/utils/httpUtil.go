package utils

import (
	"fmt"
	"io"
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

func HttpGetBody(client *http.Client, url string) (string, error) {
	resp, err := HttpGet(client, url)
	defer resp.Body.Close()

	if err != nil {
		return "", fmt.Errorf("获取响应体失败: %v", err)
	}

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close() // 确保在读取响应体失败时关闭资源
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	bodyStr := string(bodyByte)
	/*
		fmt.Println("GET请求成功，状态码：", resp.StatusCode)
		fmt.Println("响应头：", resp.Header)
		fmt.Println("响应内容：")
		fmt.Println(bodyStr)
	*/
	return bodyStr, nil
}
