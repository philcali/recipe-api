package data

type QueryParams struct {
	Limit     int    `json:"limit"`
	NextToken []byte `json:"nextToken"`
}

func (q *QueryParams) GetLimit() *int32 {
	limit := int32(100)
	if q.Limit <= 0 {
		limit = 100
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	return &limit
}

type QueryResults[T interface{}] struct {
	Items     []T    `json:"items"`
	NextToken []byte `json:"nextToken"`
}

type NextToken map[string]map[string]string
