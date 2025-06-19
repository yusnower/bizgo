package bizmgo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yusnower/gobiz/bizdb"
	"github.com/yusnower/gobiz/bizerr"
	"github.com/yusnower/gobiz/bizresult"
)

var MongoErr = bizerr.InitModuleG[struct {
	Err      bizerr.BizCode `Key:"mongoErr"`
	NotFound bizerr.BizCode `Key:"notFound"`
}]()

// NewCollection creates a new generic collection
func NewCollection[T any](coll *mongo.Collection) *Collection[T] {
	return &Collection[T]{
		coll: coll,
		ctx:  context.Background(),
	}
}

// Collection is a generic type collection wrapper
type Collection[T any] struct {
	coll *mongo.Collection
	ctx  context.Context
}

// WithCtx sets the context and returns new collection instance
func (r *Collection[T]) WithCtx(ctx context.Context) *Collection[T] {
	clone := &Collection[T]{
		coll: r.coll,
		ctx:  ctx,
	}
	return clone
}

// Create inserts a document
func (r *Collection[T]) Create(obj T) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	_, err := r.coll.InsertOne(r.ctx, obj)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, obj))
	}

	return res.SetValue(struct{}{})
}

// CreateMany inserts multiple documents
func (r *Collection[T]) CreateMany(objs []T) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	if len(objs) == 0 {
		return res.SetValue(struct{}{})
	}

	docs := make([]interface{}, len(objs))
	for i, obj := range objs {
		docs[i] = obj
	}

	_, err := r.coll.InsertMany(r.ctx, docs)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, objs))
	}

	return res.SetValue(struct{}{})
}

// Get retrieves a single document by ID
func (r *Collection[T]) Get(id interface{}) *bizresult.Result[T] {
	res := bizresult.New[T]()

	var result T

	err := r.coll.FindOne(r.ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.SetErr(MongoErr.NotFound.Wrap(err, id))
		}

		return res.SetErr(MongoErr.Err.Wrap(err, id))
	}

	return res.SetValue(result)
}

// FindOne retrieves a single document based on filter conditions
func (r *Collection[T]) FindOne(filter *bizdb.BoxFilter) *bizresult.Result[T] {
	res := bizresult.New[T]()

	var result T
	filterDoc := bizdb.FilterToMongo(filter)

	err := r.coll.FindOne(r.ctx, filterDoc).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.SetErr(MongoErr.NotFound.Wrap(err, filterDoc))
		}
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(result)
}

// FindOneAndExist retrieves a single document based on filter conditions and returns whether it exists
func (r *Collection[T]) FindOneAndExist(filter *bizdb.BoxFilter) *bizresult.Result[bizresult.WithExist[T]] {
	res := bizresult.New[bizresult.WithExist[T]]()

	var result T
	filterDoc := bizdb.FilterToMongo(filter)

	err := r.coll.FindOne(r.ctx, filterDoc).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.SetValue(bizresult.WithExist[T]{
				Data:   result,
				Exists: false,
			})
		}

		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(bizresult.WithExist[T]{
		Data:   result,
		Exists: true,
	})
}

// FindMany starts a fluent find operation
func (r *Collection[T]) FindMany(filter *bizdb.BoxFilter) *FindOptions[T] {
	return NewFindOptions(r, filter)
}

// Count calculates the number of documents matching the conditions
func (r *Collection[T]) Count(filter *bizdb.BoxFilter) *bizresult.Result[int64] {
	res := bizresult.New[int64]()

	filterDoc := bizdb.FilterToMongo(filter)

	count, err := r.coll.CountDocuments(r.ctx, filterDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(count)
}

// Update updates a single document by ID
func (r *Collection[T]) Update(id interface{}, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bson.M{"_id": id}
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.coll.UpdateOne(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, id, updateDoc))
	}

	return res.SetValue(struct{}{})
}

// UpdateOne updates a single document
func (r *Collection[T]) UpdateOne(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.coll.UpdateOne(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc, updateDoc))
	}

	return res.SetValue(struct{}{})
}

// UpdateMany updates multiple documents
func (r *Collection[T]) UpdateMany(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.coll.UpdateMany(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc, updateDoc))
	}

	return res.SetValue(struct{}{})
}

// Delete deletes a single document by ID
func (r *Collection[T]) Delete(id interface{}) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filter := bson.M{"_id": id}

	_, err := r.coll.DeleteOne(r.ctx, filter)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, id))
	}

	return res.SetValue(struct{}{})
}

// DeleteOne deletes a single document
func (r *Collection[T]) DeleteOne(filter *bizdb.BoxFilter) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)

	_, err := r.coll.DeleteOne(r.ctx, filterDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(struct{}{})
}

// DeleteMany deletes multiple document
func (r *Collection[T]) DeleteMany(filter *bizdb.BoxFilter) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)

	_, err := r.coll.DeleteMany(r.ctx, filterDoc)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(struct{}{})
}

// Aggregate executes an aggregation operation
func (r *Collection[T]) Aggregate(pipeline interface{}, results interface{}) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	cursor, err := r.coll.Aggregate(r.ctx, pipeline)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, pipeline))
	}
	defer cursor.Close(r.ctx)

	err = cursor.All(r.ctx, results)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, pipeline))
	}

	return res.SetValue(struct{}{})
}
