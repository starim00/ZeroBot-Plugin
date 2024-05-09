package aireply

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ChatGPT GPT回复类
type DeepSeek struct {
	u string
	k string
	b []string
}

// chatGPTResponseBody 响应体
type deepSeekResponseBody struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int                      `json:"created"`
	Model   string                   `json:"model"`
	Choices []map[string]interface{} `json:"choices"`
	Usage   map[string]interface{}   `json:"usage"`
}

// chatGPTRequestBody 请求体
type deepSeekRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

const (
	DeepSeekURL = "https://api.deepseek.com"
)

// NewChatGPT ...
func NewDeepSeek(u, key string, banwords ...string) *DeepSeek {
	return &DeepSeek{u: u, k: key, b: banwords}
}

// String ...
func (*DeepSeek) String() string {
	return "DeepSeek"
}

// Talk 取得带 CQ 码的回复消息
func (c *DeepSeek) Talk(_ int64, msg, _ string) string {
	replystr := deepChat(msg, c.k, c.u)
	for _, w := range c.b {
		if strings.Contains(replystr, w) {
			return "ERROR: 回复可能含有敏感内容"
		}
	}
	return replystr
}

// TalkPlain 取得回复消息
func (c *DeepSeek) TalkPlain(_ int64, msg, nickname string) string {
	return c.Talk(0, msg, nickname)
}

func deepChat(msg string, apiKey string, url string) string {
	requestBody := deepSeekRequestBody{
		Model:            "deepseek-chat",
		Prompt:           msg,
		MaxTokens:        2048,
		Temperature:      1.0,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	err := json.NewEncoder(requestData).Encode(&requestBody)
	if err != nil {
		return err.Error()
	}
	req, err := http.NewRequest("POST", url+"completions", requestData)
	if err != nil {
		return err.Error()
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer response.Body.Close()
	var deepResponseBody deepSeekResponseBody
	err = json.NewDecoder(response.Body).Decode(&deepResponseBody)
	if err != nil {
		return err.Error()
	}
	if len(deepResponseBody.Choices) > 0 {
		for _, v := range deepResponseBody.Choices {
			return fmt.Sprint(v["text"])
		}
	}
	return ""
}
