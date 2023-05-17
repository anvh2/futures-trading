package models

import "encoding/json"

type Price struct {
	Quantity float64 `json:"quantity,omitempty"`
	Entry    float64 `json:"entry,omitempty"`
	Profit   float64 `json:"profit,omitempty"`
	Loss     float64 `json:"loss,omitemty"`
}

func (p *Price) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
