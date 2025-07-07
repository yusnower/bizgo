package bizdb

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/yusnower/bizgo/bizreflect"
)

// Define update operators using a struct with biz tags
var updateOperator = bizreflect.InitStructG[struct {
	Set       string `biz:"set"`
	SqlExpr   string `biz:"sqlExpr"`
	MongoExpr string `biz:"mongoExpr"`
}]()

// UpdateOperation represents a single update operation
type updateOperation struct {
	Field string
	Op    string
	Value interface{}
	Expr  string // For custom expressions
}

// BoxUpdate structure for building update operations
type BoxUpdate struct {
	operations []updateOperation
}

// Update creates a new update builder
func Update() *BoxUpdate {
	return &BoxUpdate{
		operations: []updateOperation{},
	}
}

// Set adds a field assignment operation (field = value)
func (r *BoxUpdate) Set(field string, value interface{}) *BoxUpdate {
	r.operations = append(r.operations, updateOperation{
		Field: field,
		Op:    updateOperator.Set,
		Value: value,
	})
	return r
}

// SqlExpr adds a custom expression for SQL (field = expr)
// The expr is a string like "field + ?" or "CONCAT(field, ?)"
// and value is the parameter value for the prepared statement
func (r *BoxUpdate) SqlExpr(field, expr string, value interface{}) *BoxUpdate {
	r.operations = append(r.operations, updateOperation{
		Field: field,
		Op:    updateOperator.SqlExpr,
		Expr:  expr,
		Value: value,
	})
	return r
}

// MongoExpr adds a raw MongoDB update expression
// expr is a MongoDB update document
func (r *BoxUpdate) MongoExpr(expr bson.M) *BoxUpdate {
	r.operations = append(r.operations, updateOperation{
		Op:    updateOperator.MongoExpr,
		Value: expr,
	})
	return r
}

// UpdateToMongo converts a BoxUpdate to a MongoDB update document
func UpdateToMongo(obj *BoxUpdate) bson.M {
	if obj == nil || len(obj.operations) == 0 {
		return bson.M{}
	}

	result := bson.M{}

	setOps := bson.M{}

	// Process each operation
	for _, op := range obj.operations {
		switch op.Op {
		case updateOperator.Set:
			setOps[op.Field] = op.Value
		case updateOperator.MongoExpr:
			if expr, ok := op.Value.(bson.M); ok {
				for k, v := range expr {
					if k == "$set" {
						if fieldsMap, fieldsOk := v.(bson.M); fieldsOk {
							for field, value := range fieldsMap {
								setOps[field] = value
							}
						}
					} else {
						if _, exists := result[k]; exists {
							if existingOp, opOk := result[k].(bson.M); opOk {
								if fieldsMap, fieldsOk := v.(bson.M); fieldsOk {
									for field, value := range fieldsMap {
										existingOp[field] = value
									}
								}
							}
						} else {
							result[k] = v
						}
					}
				}
			}
		}
	}

	if len(setOps) > 0 {
		result["$set"] = setOps
	}

	return result
}
