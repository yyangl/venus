package model

type (
	ServicePoint struct {
		Type      ServiceType `json:"type"`
		Url       string      `json:"url"`
		Method    string      `json:"method"`
		ReqMethod string      `json:"req_method"`
		Ip        string      `json:"ip"`
		Port      string      `json:"port"`
	}

	ServiceType uint8
)

const (
	HttpService ServiceType = iota
	RpcService
)
