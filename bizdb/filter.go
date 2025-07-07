package bizdb

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/yusnower/bizgo/bizreflect"
)

var operator = bizreflect.InitStructG[struct {
	Eq       string `biz:"eq"`
	NotEq    string `biz:"ne"`
	Gt       string `biz:"gt"`
	Gte      string `biz:"gte"`
	Lt       string `biz:"lt"`
	Lte      string `biz:"lte"`
	In       string `biz:"in"`
	NotIn    string `biz:"notin"`
	Like     string `biz:"like"`
	Or       string `biz:"or"`
	And      string `biz:"and"`
	RawSql   string `biz:"rawSql"`
	RawMongo string `biz:"rawMongo"`
}]()

// filterItem represents a single filter filterItem
type filterItem struct {
	Field    string
	Op       string
	Value    interface{}
	Children []*BoxFilter // For composite conditions like OR, AND
}

// BoxFilter structure for building query conditions
type BoxFilter struct {
	conditions []filterItem
}

// Filter creates a new filter instance
func Filter() *BoxFilter {
	return &BoxFilter{
		conditions: []filterItem{},
	}
}

// Eq adds an equality filterItem
func (r *BoxFilter) Eq(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Eq,
		Value: value,
	})
	return r
}

// NotEq adds a not-equal filterItem
func (r *BoxFilter) NotEq(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.NotEq,
		Value: value,
	})
	return r
}

// Gt adds a greater-than filterItem
func (r *BoxFilter) Gt(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Gt,
		Value: value,
	})
	return r
}

// Gte adds a greater-than-or-equal filterItem
func (r *BoxFilter) Gte(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Gte,
		Value: value,
	})
	return r
}

// Lt adds a less-than filterItem
func (r *BoxFilter) Lt(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Lt,
		Value: value,
	})
	return r
}

// Lte adds a less-than-or-equal filterItem
func (r *BoxFilter) Lte(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Lte,
		Value: value,
	})
	return r
}

// In adds an in-array filterItem
func (r *BoxFilter) In(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.In,
		Value: value,
	})
	return r
}

// NotIn adds a not-in-array filterItem
func (r *BoxFilter) NotIn(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.NotIn,
		Value: value,
	})
	return r
}

// Like adds a pattern matching filterItem
func (r *BoxFilter) Like(field string, value interface{}) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Field: field,
		Op:    operator.Like,
		Value: value,
	})
	return r
}

// Or adds a composite OR filterItem with the provided filters
func (r *BoxFilter) Or(filters ...*BoxFilter) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Op:       operator.Or,
		Children: filters,
	})
	return r
}

// And adds a composite AND filterItem with the provided filters
func (r *BoxFilter) And(filters ...*BoxFilter) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Op:       operator.And,
		Children: filters,
	})
	return r
}

// Mongo adds a raw MongoDB filter condition to the filter
// filter: MongoDB query document (bson.M)
func (r *BoxFilter) Mongo(filter bson.M) *BoxFilter {
	r.conditions = append(r.conditions, filterItem{
		Op:    operator.RawMongo,
		Value: filter,
	})

	return r
}

// FilterToMongo converts a BoxFilter to a MongoDB query document
func FilterToMongo(r *BoxFilter) bson.M {
	// Handle nil filter
	if r == nil {
		return bson.M{}
	}

	result := bson.M{}

	if len(r.conditions) == 0 {
		return result
	}

	// Collect complex conditions that require $and/$or operators
	var complexConditions []bson.M

	for _, condition := range r.conditions {
		switch condition.Op {
		case operator.RawMongo:
			if mongoFilter, ok := condition.Value.(bson.M); ok {
				for k, v := range mongoFilter {
					if k == "$and" {
						if andArray, ok := v.([]interface{}); ok {
							for _, item := range andArray {
								if itemMap, ok := item.(bson.M); ok {
									complexConditions = append(complexConditions, itemMap)
								} else if itemMap, ok := item.(map[string]interface{}); ok {
									complexConditions = append(complexConditions, itemMap)
								}
							}
						} else if andArray, ok := v.([]bson.M); ok {
							complexConditions = append(complexConditions, andArray...)
						}
					} else if k == "$or" {
						// Keep $or as a complex condition
						result[k] = v
					} else {
						result[k] = v
					}
				}
			}
		case operator.Eq:
			result[condition.Field] = condition.Value
		case operator.NotEq:
			result[condition.Field] = bson.M{"$ne": condition.Value}
		case operator.Gt:
			result[condition.Field] = bson.M{"$gt": condition.Value}
		case operator.Gte:
			result[condition.Field] = bson.M{"$gte": condition.Value}
		case operator.Lt:
			result[condition.Field] = bson.M{"$lt": condition.Value}
		case operator.Lte:
			result[condition.Field] = bson.M{"$lte": condition.Value}
		case operator.In:
			result[condition.Field] = bson.M{"$in": condition.Value}
		case operator.NotIn:
			result[condition.Field] = bson.M{"$nin": condition.Value}
		case operator.Like:
			// MongoDB uses regex for LIKE operations
			if str, ok := condition.Value.(string); ok {
				pattern := strings.Replace(str, "%", ".*", -1)
				result[condition.Field] = bson.M{"$regex": pattern, "$options": "i"}
			}
		case operator.Or:
			var orConditions []bson.M
			for _, child := range condition.Children {
				orConditions = append(orConditions, FilterToMongo(child))
			}
			if len(orConditions) > 0 {
				// Add directly to the result if there's no other OR condition
				if _, exists := result["$or"]; !exists {
					result["$or"] = orConditions
				} else {
					// If we already have an $or, we need to use $and to combine them
					complexConditions = append(complexConditions, bson.M{"$or": orConditions})
				}
			}
		case operator.And:
			var andConditions []bson.M
			for _, child := range condition.Children {
				childFilter := FilterToMongo(child)
				// If the child filter is not empty, add it
				if len(childFilter) > 0 {
					// If it's a simple filter with only one condition, flatten it
					if len(childFilter) == 1 && !hasLogicalOperator(childFilter) {
						for k, v := range childFilter {
							result[k] = v
						}
					} else {
						andConditions = append(andConditions, childFilter)
					}
				}
			}

			// Only add non-empty andConditions
			if len(andConditions) > 0 {
				complexConditions = append(complexConditions, andConditions...)
			}
		}
	}

	// Add complex conditions if needed
	if len(complexConditions) > 0 {
		// If we have exactly one complex condition and no other conditions,
		// and it doesn't conflict with existing fields, we can flatten it
		if len(complexConditions) == 1 && len(result) == 0 {
			return complexConditions[0]
		}

		// If we already have $and conditions, append to them
		if existingAnd, exists := result["$and"]; exists {
			switch v := existingAnd.(type) {
			case []bson.M:
				result["$and"] = append(v, complexConditions...)
			case []interface{}:
				// Convert existing conditions to bson.M if needed
				newAnd := make([]bson.M, 0, len(v)+len(complexConditions))
				for _, item := range v {
					if itemMap, ok := item.(bson.M); ok {
						newAnd = append(newAnd, itemMap)
					} else if itemMap, ok := item.(map[string]interface{}); ok {
						newAnd = append(newAnd, itemMap)
					}
				}
				newAnd = append(newAnd, complexConditions...)
				result["$and"] = newAnd
			default:
				// If it's something else, create a new $and array
				result["$and"] = complexConditions
			}
		} else {
			result["$and"] = complexConditions
		}
	}

	return result
}

// hasLogicalOperator checks if a bson.M contains any logical operator ($and, $or, etc.)
func hasLogicalOperator(doc bson.M) bool {
	for k := range doc {
		if strings.HasPrefix(k, "$") {
			return true
		}
	}
	return false
}
