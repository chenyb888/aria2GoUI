package aria2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client aria2 RPC 客户端
type Client struct {
	rpcURL string
	token  string
	client *http.Client
}

// NewClient 创建新的 aria2 RPC 客户端
func NewClient(host string, port int, token string, protocol string, path string) *Client {
	rpcURL := fmt.Sprintf("%s://%s:%d%s", protocol, host, port, path)
	return &Client{
		rpcURL: rpcURL,
		token:  token,
		client: &http.Client{},
	}
}

// RPCRequest RPC 请求结构
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      string        `json:"id"`
}

// RPCResponse RPC 响应结构
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
	ID      string          `json:"id"`
}

// RPCError RPC 错误结构
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Version aria2 版本信息
type Version struct {
	Version     string   `json:"version"`
	EnabledFeatures []string `json:"enabledFeatures"`
}

// TellStatus 任务状态
type TellStatus struct {
	GID           string            `json:"gid"`
	Status        string            `json:"status"`
	TotalLength   string            `json:"totalLength"`
	CompletedLength string          `json:"completedLength"`
	UploadLength  string            `json:"uploadLength"`
	Bitfield      string            `json:"bitfield"`
	DownloadSpeed string            `json:"downloadSpeed"`
	UploadSpeed   string            `json:"uploadSpeed"`
	InfoHash      string            `json:"infoHash"`
	NumSeeders    int               `json:"numSeeders"`
	Seeder        string            `json:"seeder"`
	PieceLength   string            `json:"pieceLength"`
	NumPieces     int               `json:"numPieces"`
	Connections   int               `json:"connections"`
	ErrorCode     string            `json:"errorCode"`
	ErrorMessage  string            `json:"errorMessage"`
	FollowedBy    []string          `json:"followedBy"`
	BelongsTo     string            `json:"belongsTo"`
	Dir           string            `json:"dir"`
	Files         []FileInfo        `json:"files"`
	Bittorrent    *BittorrentInfo   `json:"bittorrent"`
	VerifiedLength string           `json:"verifiedLength"`
	VerifyIntegrityPending string   `json:"verifyIntegrityPending"`
}

// FileInfo 文件信息
type FileInfo struct {
	Index    string   `json:"index"`
	Path     string   `json:"path"`
	Length   string   `json:"length"`
	CompletedLength string `json:"completedLength"`
	Selected string   `json:"selected"`
	URIs     []URI    `json:"uris"`
}

// URI URI 信息
type URI struct {
	Status string `json:"status"`
	URI    string `json:"uri"`
}

// BittorrentInfo 种子信息
type BittorrentInfo struct {
	AnnounceList [][]string `json:"announceList"`
	Comment      string     `json:"comment"`
	CreationDate int64      `json:"creationDate"`
	Mode         string     `json:"mode"`
	Info         struct {
		Name string `json:"name"`
	} `json:"info"`
}

// GetVersion 获取 aria2 版本信息
func (c *Client) GetVersion() (*Version, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.getVersion",
		Params:  []interface{}{"token:" + c.token},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var version Version
	if err := json.Unmarshal(response.Result, &version); err != nil {
		return nil, err
	}

	return &version, nil
}

// TellActive 获取活动任务列表
func (c *Client) TellActive() ([]TellStatus, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.tellActive",
		Params:  []interface{}{"token:" + c.token},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var tasks []TellStatus
	if err := json.Unmarshal(response.Result, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// TellWaiting 获取等待任务列表
func (c *Client) TellWaiting(offset int, num int) ([]TellStatus, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.tellWaiting",
		Params:  []interface{}{"token:" + c.token, offset, num},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var tasks []TellStatus
	if err := json.Unmarshal(response.Result, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// TellStopped 获取已停止任务列表
func (c *Client) TellStopped(offset int, num int) ([]TellStatus, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.tellStopped",
		Params:  []interface{}{"token:" + c.token, offset, num},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var tasks []TellStatus
	if err := json.Unmarshal(response.Result, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// AddURI 添加下载任务
func (c *Client) AddURI(uris []string, options map[string]interface{}) (string, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.addUri",
		Params:  []interface{}{"token:" + c.token, uris, options},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return "", err
	}

	if response.Error != nil {
		return "", fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var gid string
	if err := json.Unmarshal(response.Result, &gid); err != nil {
		return "", err
	}

	return gid, nil
}

// Pause 暂停任务
func (c *Client) Pause(gid string) error {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.pause",
		Params:  []interface{}{"token:" + c.token, gid},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

// Unpause 恢复任务
func (c *Client) Unpause(gid string) error {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.unpause",
		Params:  []interface{}{"token:" + c.token, gid},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

// Remove 删除任务
func (c *Client) Remove(gid string) error {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.remove",
		Params:  []interface{}{"token:" + c.token, gid},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

// PauseAll 暂停所有任务
func (c *Client) PauseAll() error {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.pauseAll",
		Params:  []interface{}{"token:" + c.token},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

// UnpauseAll 恢复所有任务
func (c *Client) UnpauseAll() error {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.unpauseAll",
		Params:  []interface{}{"token:" + c.token},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	return nil
}

// GetGlobalStat 获取全局统计信息
func (c *Client) GetGlobalStat() (map[string]string, error) {
	request := RPCRequest{
		JSONRPC: "2.0",
		Method:  "aria2.getGlobalStat",
		Params:  []interface{}{"token:" + c.token},
		ID:      "1",
	}

	response, err := c.sendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	var stats map[string]string
	if err := json.Unmarshal(response.Result, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// sendRequest 发送 RPC 请求
func (c *Client) sendRequest(request RPCRequest) (*RPCResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.rpcURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResponse RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return nil, err
	}

	return &rpcResponse, nil
}