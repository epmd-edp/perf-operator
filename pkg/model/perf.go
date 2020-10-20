package model

type DataSource struct {
	Id     int                    `json:"id"`
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Active string                 `json:"active"`
	Config map[string]interface{} `json:"config"`
}
