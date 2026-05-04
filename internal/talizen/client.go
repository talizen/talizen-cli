package talizen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) do(ctx context.Context, method string, path string, query url.Values, body any, out any) error {
	var reader io.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(bs)
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		_ = json.Unmarshal(bs, &apiErr)
		if apiErr.Message != "" {
			return fmt.Errorf("%s %s: %s", method, path, apiErr.Message)
		}
		return fmt.Errorf("%s %s: status %d", method, path, resp.StatusCode)
	}

	if out == nil || len(bs) == 0 {
		return nil
	}

	err = json.Unmarshal(bs, out)
	if err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	return nil
}

type CLIAuthSession struct {
	Code      string `json:"code"`
	VerifyURL string `json:"verify_url"`
	ExpiresIn int    `json:"expires_in"`
}

type CLIAuthSessionResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

func (c *Client) CreateCLIAuthSession(ctx context.Context, webURL string) (CLIAuthSession, error) {
	var ret CLIAuthSession
	err := c.do(ctx, http.MethodPost, "/api/u/cli/auth/session", nil, map[string]any{
		"web_url": strings.TrimSpace(webURL),
	}, &ret)
	if err != nil {
		return CLIAuthSession{}, err
	}

	return ret, nil
}

func (c *Client) GetCLIAuthSession(ctx context.Context, code string) (CLIAuthSessionResult, error) {
	var ret CLIAuthSessionResult
	err := c.do(ctx, http.MethodGet, "/api/u/cli/auth/session/"+url.PathEscape(code), nil, nil, &ret)
	if err != nil {
		return CLIAuthSessionResult{}, err
	}

	return ret, nil
}

type Site struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SiteList []Site `json:"site_list"`
}

type ProjectListResponse struct {
	Total int       `json:"total"`
	List  []Project `json:"list"`
}

func (c *Client) GetProjectList(ctx context.Context) (ProjectListResponse, error) {
	var ret ProjectListResponse
	err := c.do(ctx, http.MethodGet, "/api/u/project_list", nil, nil, &ret)
	if err != nil {
		return ProjectListResponse{}, err
	}

	return ret, nil
}

type SystemInfo struct {
	SelfAPIHost string `json:"self_api_host"`
}

func (c *Client) GetSystemInfo(ctx context.Context) (SystemInfo, error) {
	var ret SystemInfo
	err := c.do(ctx, http.MethodGet, "/api/u/system/info", nil, nil, &ret)
	if err != nil {
		return SystemInfo{}, err
	}

	return ret, nil
}

type File struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Body     string `json:"body"`
	Hash     string `json:"hash"`
	Readonly bool   `json:"readonly"`
	IsDir    bool   `json:"is_dir"`
}

type FileListResponse struct {
	List []File `json:"list"`
}

func (c *Client) GetFileList(ctx context.Context, projectID string, siteID string) (FileListResponse, error) {
	var ret FileListResponse
	path := fmt.Sprintf("/api/u/project/%s/site/%s/file_list", url.PathEscape(projectID), url.PathEscape(siteID))
	err := c.do(ctx, http.MethodGet, path, nil, nil, &ret)
	if err != nil {
		return FileListResponse{}, err
	}

	return ret, nil
}

type SiteActionFileSpec struct {
	ID   string  `json:"id,omitempty"`
	Path *string `json:"path,omitempty"`
	Body *string `json:"body,omitempty"`
}

type SiteActionChange struct {
	Action string             `json:"action"`
	File   SiteActionFileSpec `json:"file"`
}

type SiteActionResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		Total   int `json:"total"`
		Success int `json:"success"`
		Failed  int `json:"failed"`
	} `json:"result"`
}

func (c *Client) DoSiteAction(ctx context.Context, projectID string, siteID string, clientID string, changes []SiteActionChange) (SiteActionResponse, error) {
	var ret SiteActionResponse
	path := fmt.Sprintf("/api/u/project/%s/site/%s/site_action", url.PathEscape(projectID), url.PathEscape(siteID))
	body := map[string]any{
		"client_id": clientID,
		"changes":   changes,
	}
	err := c.do(ctx, http.MethodPost, path, nil, body, &ret)
	if err != nil {
		return SiteActionResponse{}, err
	}
	if ret.Result.Failed > 0 {
		return ret, fmt.Errorf("site action partially failed: %d/%d", ret.Result.Failed, ret.Result.Total)
	}

	return ret, nil
}

func (c *Client) PublishSite(ctx context.Context, projectID string, siteID string, note string) error {
	path := fmt.Sprintf("/api/u/project/%s/site/%s/publish", url.PathEscape(projectID), url.PathEscape(siteID))
	body := map[string]any{
		"note": strings.TrimSpace(note),
	}

	return c.do(ctx, http.MethodPost, path, nil, body, nil)
}

func StringPtr(v string) *string {
	return &v
}
