package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client API 客户端
// 决策：使用标准 net/http，简单无依赖。可选 resty（更丰富的功能）
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
}

// New 创建新客户端
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetToken 设置认证 Token
func (c *Client) SetToken(token string) {
	c.token = token
}

// Do 发送 HTTP 请求
func (c *Client) Do(method, path string, body interface{}, result interface{}) error {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// 设置 Headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// 发送请求
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// Get 发送 GET 请求
func (c *Client) Get(path string, result interface{}) error {
	return c.Do(http.MethodGet, path, nil, result)
}

// Post 发送 POST 请求
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.Do(http.MethodPost, path, body, result)
}

// Delete 发送 DELETE 请求
func (c *Client) Delete(path string, result interface{}) error {
	return c.Do(http.MethodDelete, path, nil, result)
}

// Auth 返回认证 API 客户端
func (c *Client) Auth() *AuthAPI {
	return &AuthAPI{client: c}
}

// AuthAPI 认证相关 API
type AuthAPI struct {
	client *Client
}

// Register 注册新 Agent
func (a *AuthAPI) Register(req *RegisterRequest) (*RegisterResponse, error) {
	var resp RegisterResponse
	if err := a.client.Post("/api/v1/auth/register", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Challenge 获取认证挑战
func (a *AuthAPI) Challenge(publicKey string) (*ChallengeResponse, error) {
	req := &ChallengeRequest{PublicKey: publicKey}
	var resp ChallengeResponse
	if err := a.client.Post("/api/v1/auth/challenge", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Token 获取访问 Token
func (a *AuthAPI) Token(req *TokenRequest) (*TokenResponse, error) {
	var resp TokenResponse
	if err := a.client.Post("/api/v1/auth/token", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Commands 获取命令列表
func (c *Client) Commands() (*CommandsResponse, error) {
	var resp CommandsResponse
	if err := c.Get("/api/v1/commands", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
