package model

type (
	ServicePoint struct {
		Name    string      `json:"name"`
		Type    ServiceType `json:"type"`
		Ip      string      `json:"ip"`
		Port    int         `json:"port"`
		Methods []Methods   `json:"methods"`
		Id      string      `json:"id"`
	}

	ServiceType uint8

	Methods struct {
		Method    string `json:"method"`
		ReqMethod string `json:"req_method"`
		Url       string `json:"url"`
		Version   string `json:"version"`
	}
)

const (
	HttpService ServiceType = iota
	RpcService
)
