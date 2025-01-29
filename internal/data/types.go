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

type Repository[T interface{}, I interface{}] interface {
	Get(accountId string, itemId string) (T, error)
	Create(accountId string, input I) (T, error)
	CreateWithItemId(accountId string, input I, itemId string) (T, error)
	Update(accountId string, itemId string, input I) (T, error)
	List(accountId string, params QueryParams) (QueryResults[T], error)
	ListByIndex(accountId string, indexName string, params QueryParams) (QueryResults[T], error)
	Delete(accountId string, itemId string) error
}
