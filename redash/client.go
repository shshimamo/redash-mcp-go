package redash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client は Redash API クライアント
type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// NewClient は新しい Redash クライアントを作成
func NewClient(baseURL, apiKey string, noProxy bool) *Client {
	// HTTP トランスポートの設定
	transport := http.DefaultTransport.(*http.Transport).Clone()

	// noProxy が true の場合、プロキシを無効化
	// プロキシ環境で内部 Redash に接続する場合に使用
	if noProxy {
		transport.Proxy = nil
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// Redash API のレスポンス型定義

// QueryExecuteResponse はクエリ実行結果（2パターンある）
// パターン1: キャッシュがある場合 {"query_result": {...}}
// パターン2: 新規実行の場合 {"job": {...}}
type QueryExecuteResponse struct {
	Job         *QueryJob    `json:"job,omitempty"`
	QueryResult *QueryResult `json:"query_result,omitempty"`
}

type QueryJob struct {
	ID          string       `json:"id"`
	Status      int          `json:"status"` // 1: pending, 2: started, 3: success, 4: failure
	Error       string       `json:"error,omitempty"`
	QueryResult *QueryResult `json:"query_result,omitempty"`
}

type QueryResult struct {
	ID   int             `json:"id"`
	Data json.RawMessage `json:"data"`
}

// QueryResultData はクエリ結果のデータ部分
type QueryResultData struct {
	Columns []Column        `json:"columns"`
	Rows    json.RawMessage `json:"rows"`
}

type Column struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Type         string `json:"type"`
}

// Query はクエリのメタデータ
type Query struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Query        string `json:"query"`
	DataSourceID int    `json:"data_source_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Dashboard はダッシュボードのメタデータ
type Dashboard struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Widgets   []Widget  `json:"widgets"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// Widget はダッシュボード内のウィジェット
type Widget struct {
	ID            int            `json:"id"`
	Visualization *Visualization `json:"visualization,omitempty"`
}

// Visualization はウィジェット内のビジュアライゼーション
type Visualization struct {
	ID    int          `json:"id"`
	Name  string       `json:"name"`
	Type  string       `json:"type"`
	Query *WidgetQuery `json:"query,omitempty"`
}

// WidgetQuery はビジュアライゼーションに紐づくクエリ
type WidgetQuery struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Query        string `json:"query"`
	DataSourceID int    `json:"data_source_id"`
}

// Alert はアラートのメタデータ
type Alert struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Query     *Query                 `json:"query,omitempty"`
	State     string                 `json:"state"`
	Options   map[string]interface{} `json:"options"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// ExecuteQuery は保存済みクエリを実行
// query_id: 実行するクエリのID
// parameters: クエリパラメータ（オプション）
func (c *Client) ExecuteQuery(queryID int, parameters map[string]interface{}) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/queries/%d/results", c.BaseURL, queryID)

	// パラメータがある場合は JSON エンコード
	var body io.Reader
	if len(parameters) > 0 {
		paramsJSON := map[string]interface{}{
			"parameters": parameters,
		}
		data, err := json.Marshal(paramsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameters: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// レスポンスをパース
	var result QueryExecuteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// パターン1: キャッシュがある場合は直接結果を返す
	if result.QueryResult != nil {
		return result.QueryResult.Data, nil
	}

	// パターン2: ジョブの完了を待つ
	if result.Job != nil {
		return c.waitForJob(result.Job.ID)
	}

	return nil, fmt.Errorf("unexpected response format: no query_result or job found")
}

// ExecuteAdhocQuery はアドホッククエリを実行
// query: 実行する SQL
// dataSourceID: データソースID
func (c *Client) ExecuteAdhocQuery(query string, dataSourceID int) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/query_results", c.BaseURL)

	reqBody := map[string]interface{}{
		"query":          query,
		"data_source_id": dataSourceID,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// レスポンスをパース
	var result QueryExecuteResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// パターン1: キャッシュがある場合は直接結果を返す
	if result.QueryResult != nil {
		return result.QueryResult.Data, nil
	}

	// パターン2: ジョブの完了を待つ
	if result.Job != nil {
		return c.waitForJob(result.Job.ID)
	}

	return nil, fmt.Errorf("unexpected response format: no query_result or job found")
}

// waitForJob はジョブの完了を待機してクエリ結果を返す
func (c *Client) waitForJob(jobID string) (json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/jobs/%s", c.BaseURL, jobID)

	// 最大30回、1秒間隔でポーリング
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create job status request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get job status: %w", err)
		}

		var job QueryJob
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode job status: %w", err)
		}
		resp.Body.Close()

		switch job.Status {
		case 3: // Success
			if job.QueryResult != nil {
				return job.QueryResult.Data, nil
			}
			return nil, fmt.Errorf("query succeeded but no result data")
		case 4: // Failure
			return nil, fmt.Errorf("query failed: %s", job.Error)
		case 1, 2: // Pending or Started
			continue
		}
	}

	return nil, fmt.Errorf("query timeout: job did not complete in 30 seconds")
}

// GetQuery はクエリのメタデータを取得
func (c *Client) GetQuery(queryID int) (*Query, error) {
	url := fmt.Sprintf("%s/api/queries/%d", c.BaseURL, queryID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var query Query
	if err := json.NewDecoder(resp.Body).Decode(&query); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &query, nil
}

// GetDashboard はダッシュボードのメタデータを取得
func (c *Client) GetDashboard(dashboardID int) (*Dashboard, error) {
	url := fmt.Sprintf("%s/api/dashboards/%d", c.BaseURL, dashboardID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var dashboard Dashboard
	if err := json.NewDecoder(resp.Body).Decode(&dashboard); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &dashboard, nil
}

// GetAlert はアラートのメタデータを取得
func (c *Client) GetAlert(alertID int) (*Alert, error) {
	url := fmt.Sprintf("%s/api/alerts/%d", c.BaseURL, alertID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Key %s", c.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var alert Alert
	if err := json.NewDecoder(resp.Body).Decode(&alert); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &alert, nil
}
