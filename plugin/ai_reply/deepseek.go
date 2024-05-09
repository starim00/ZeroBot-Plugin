package aireply

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// DeepSeek 回复类
type DeepSeek struct {
	u string
	k string
	b []string
}

// deepSeekResponseBody 响应体
type deepSeekResponseBody struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// deepSeekRequestBody 请求体
type deepSeekRequestBody struct {
	Model    string `json:"model"`
	Messages []struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	} `json:"messages"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

const (
	DeepSeekURL = "https://api.deepseek.com/chat/"
)

// NewDeepSeek ...
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
		Model: "deepseek-chat",
		Messages: []struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		}{
			{
				Content: msg,
				Role:    "user",
			},
		},
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
		return deepResponseBody.Choices[0].Message.Content
	}
	return ""
}
