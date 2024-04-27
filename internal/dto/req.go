package dto

import "time"

type IpReqDb struct {
	MaxReq     int     `json:"max_req"`
	TimeWindow int64   `json:"time_window"`
	Req        []int64 `json:"req"`
}

type IpReq struct {
	IP        string
	TimeAdded time.Time
}

type IpAllow struct {
	Allow bool
}
