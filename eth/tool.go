package eth

import (
	"encoding/json"
	"fmt"
)

type NodeInfo struct {
	Addr       string
	EnableTLS  bool
	CertPath   string
	CommonName string
	AccessCert string
	AccessKey  string
}

// CompileResult is packaged compile contract result
type CompileResult struct {
	Abi   []string
	Bin   []string
	Types []string
}

// EthError - ethereum error
type EthError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err EthError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type ethResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *EthError       `json:"error"`
}

type ethRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}
