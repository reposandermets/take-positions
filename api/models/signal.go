package models

type Signal struct {
	Ticker   string `json:"ticker"`
	Exchange string `json:"exchange"`
	Signal   string `json:"signal"`
	Type     string `json:"type"`
}

type RequestContext struct {
	Ticker   string
	Exchange string
	Signal   string
	Type     string
}
