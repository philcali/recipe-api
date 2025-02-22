package audits

import "time"

type Audit struct {
	Id         string    `json:"id"`
	Message    string    `json:"message"`
	CreateTime time.Time `json:"createTime"`
}
