package audits

import "time"

type Audit struct {
	Id           string                  `json:"id"`
	Action       string                  `json:"action"`
	ResourceType string                  `json:"resourceType"`
	ResourceId   string                  `json:"resourceId"`
	NewValues    *map[string]interface{} `json:"newValues"`
	OldValues    *map[string]interface{} `json:"oldValues"`
	CreateTime   time.Time               `json:"createTime"`
	UpdateTime   time.Time               `json:"updateTime"`
}
