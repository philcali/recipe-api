package mealdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"philcali.me/recipes/internal/data"
	"philcali.me/recipes/internal/provider"
	"philcali.me/recipes/internal/routes/recipes"
	"philcali.me/recipes/internal/routes/util"
)

type MealAPI struct {
	Version string
	Token   string
	Client  *http.Client
}

func _apiRequent(mc *MealAPI, resource string, params map[string]string) ([]byte, error) {
	queryParams := make([]string, len(params))
	for k, v := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, v))
	}
	formatParams := ""
	if len(queryParams) > 0 {
		formatParams = "?" + strings.Join(queryParams, "&")
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://themealdb.com/api/json/%s/%s/%s.php%s", mc.Version, mc.Token, resource, formatParams), nil)
	if err != nil {
		return nil, err
	}
	resp, err := mc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func _convertQuery(query QueryResponse) data.QueryResults[recipes.Recipe] {
	return util.ConvertQueryResults(data.QueryResults[Meal]{
		Items: query.Meals,
	}, ToRecipe)
}

func _convertFilter(query FilterResponse) data.QueryResults[recipes.Recipe] {
	return util.ConvertQueryResults(data.QueryResults[FilteredMeal]{
		Items: query.Meals,
	}, ConvertCategoryToRecipe)
}

func _queryRequest(mc *MealAPI, resource string, params map[string]string) (data.QueryResults[recipes.Recipe], error) {
	body, err := _apiRequent(mc, resource, params)
	if err != nil {
		return data.QueryResults[recipes.Recipe]{}, err
	}
	var query QueryResponse
	if err := json.Unmarshal(body, &query); err != nil {
		return data.QueryResults[recipes.Recipe]{}, err
	}

	return _convertQuery(query), nil
}

func (mc *MealAPI) Random() (data.QueryResults[recipes.Recipe], error) {
	return _queryRequest(mc, "random", map[string]string{})
}

func (mc *MealAPI) Filter(input provider.FilterInput) (data.QueryResults[recipes.Recipe], error) {
	params := make(map[string]string, 1)
	if input.Category != nil {
		params["c"] = *input.Category
	}
	if input.Area != nil {
		params["a"] = *input.Area
	}
	if input.MainIngredient != nil {
		params["i"] = *input.MainIngredient
	}
	body, err := _apiRequent(mc, "filter", params)
	if err != nil {
		return data.QueryResults[recipes.Recipe]{}, err
	}
	var filter FilterResponse
	if err := json.Unmarshal(body, &filter); err != nil {
		return data.QueryResults[recipes.Recipe]{}, err
	}
	return _convertFilter(filter), nil
}

func (mc *MealAPI) Lookup(id string) (data.QueryResults[recipes.Recipe], error) {
	return _queryRequest(mc, "lookup", map[string]string{
		"i": id,
	})
}

func (mc *MealAPI) Search(text string) (data.QueryResults[recipes.Recipe], error) {
	return _queryRequest(mc, "search", map[string]string{
		"s": text,
	})
}

func NewDefaultMealClient() provider.RecipeProvider {
	return &MealAPI{
		Version: "v1",
		Token:   "1",
		Client:  &http.Client{},
	}
}
