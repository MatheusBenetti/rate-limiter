package dto

import "time"

type ApiKeyReqDb struct {
	MaxReq     int     `json:"max_req"`
	TimeWindow int64   `json:"time_window"`
	Req        []int64 `json:"req"`
}

type ApiKeyReq struct {
	Value     string
	TimeAdded time.Time
}

type Input struct {
	MaxReq        int   `json:"max_req"`
	TimeWindow    int64 `json:"time_window"`
	BlockDuration int64 `json:"block_duration"`
}

type Output struct {
	Api_Key string `json:"api_key"`
}

type ApiKeyAllow struct {
	Allow bool
}
