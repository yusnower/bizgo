package bizmgo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/yusnower/bizgo/bizdb"
	"github.com/yusnower/bizgo/bizerr"
	"github.com/yusnower/bizgo/bizresult"
)

var MongoErr = bizerr.InitModuleG[struct {
	Err      bizerr.BizCode `Key:"mongoErr"`
	NotFound bizerr.BizCode `Key:"notFound"`
}]()

// NewCollection creates a new generic collection
func NewCollection[T any](db *mongo.Database, collection string) *Collection[T] {
	return &Collection[T]{
		db:   db,
		coll: collection,
		ctx:  context.Background(),
	}
}

// Collection is a generic type collection wrapper
type Collection[T any] struct {
	db   *mongo.Database
	coll string
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

	_, err := r.getCollection().InsertOne(r.ctx, obj)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, obj))
	}

	return res.Ok(struct{}{})
}

// CreateMany inserts multiple documents
func (r *Collection[T]) CreateMany(objs []T) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	if len(objs) == 0 {
		return res.Ok(struct{}{})
	}

	docs := make([]interface{}, len(objs))
	for i, obj := range objs {
		docs[i] = obj
	}

	_, err := r.getCollection().InsertMany(r.ctx, docs)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, objs))
	}

	return res.Ok(struct{}{})
}

// Get retrieves a single document by ID
func (r *Collection[T]) Get(id interface{}) *bizresult.Result[T] {
	res := bizresult.New[T]()

	var result T

	err := r.getCollection().FindOne(r.ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.Err(MongoErr.NotFound.Wrap(err, id))
		}

		return res.Err(MongoErr.Err.Wrap(err, id))
	}

	return res.Ok(result)
}

// FindOne retrieves a single document based on filter conditions
func (r *Collection[T]) FindOne(filter *bizdb.BoxFilter) *bizresult.Result[T] {
	res := bizresult.New[T]()

	var result T
	filterDoc := bizdb.FilterToMongo(filter)

	err := r.getCollection().FindOne(r.ctx, filterDoc).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.Err(MongoErr.NotFound.Wrap(err, filterDoc))
		}
		return res.Err(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.Ok(result)
}

// FindOneAndExist retrieves a single document based on filter conditions and returns whether it exists
func (r *Collection[T]) FindOneAndExist(filter *bizdb.BoxFilter) *bizresult.Result[bizresult.WithExist[T]] {
	res := bizresult.New[bizresult.WithExist[T]]()

	var result T
	filterDoc := bizdb.FilterToMongo(filter)

	err := r.getCollection().FindOne(r.ctx, filterDoc).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return res.Ok(bizresult.WithExist[T]{
				Data:   result,
				Exists: false,
			})
		}

		return res.Err(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.Ok(bizresult.WithExist[T]{
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

	count, err := r.getCollection().CountDocuments(r.ctx, filterDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.Ok(count)
}

// Update updates a single document by ID
func (r *Collection[T]) Update(id interface{}, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bson.M{"_id": id}
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.getCollection().UpdateOne(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, id, updateDoc))
	}

	return res.Ok(struct{}{})
}

// UpdateOne updates a single document
func (r *Collection[T]) UpdateOne(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.getCollection().UpdateOne(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, filterDoc, updateDoc))
	}

	return res.Ok(struct{}{})
}

// UpdateMany updates multiple documents
func (r *Collection[T]) UpdateMany(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	_, err := r.getCollection().UpdateMany(r.ctx, filterDoc, updateDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, filterDoc, updateDoc))
	}

	return res.Ok(struct{}{})
}

// Delete deletes a single document by ID
func (r *Collection[T]) Delete(id interface{}) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filter := bson.M{"_id": id}

	_, err := r.getCollection().DeleteOne(r.ctx, filter)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, id))
	}

	return res.Ok(struct{}{})
}

// DeleteOne deletes a single document
func (r *Collection[T]) DeleteOne(filter *bizdb.BoxFilter) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)

	_, err := r.getCollection().DeleteOne(r.ctx, filterDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.Ok(struct{}{})
}

// DeleteMany deletes multiple document
func (r *Collection[T]) DeleteMany(filter *bizdb.BoxFilter) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)

	_, err := r.getCollection().DeleteMany(r.ctx, filterDoc)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.Ok(struct{}{})
}

// Aggregate executes an aggregation operation
func (r *Collection[T]) Aggregate(pipeline interface{}, results interface{}) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	cursor, err := r.getCollection().Aggregate(r.ctx, pipeline)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, pipeline))
	}
	defer cursor.Close(r.ctx)

	err = cursor.All(r.ctx, results)
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, pipeline))
	}

	return res.Ok(struct{}{})
}

func (r *Collection[T]) getCollection() *mongo.Collection {
	return r.db.Collection(r.coll)
}

func (r *Collection[T]) Raw(fn func(ctx context.Context, coll *mongo.Collection) error) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	err := fn(r.ctx, r.getCollection())
	if err != nil {
		return res.Err(MongoErr.Err.Wrap(err, r.coll))
	}

	return res.Ok(struct{}{})
}

func (r *Collection[T]) FindAndUpdate(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate) *bizresult.Result[T] {
	res := bizresult.New[T]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	var result T
	if err := r.getCollection().FindOneAndUpdate(r.ctx, filterDoc, updateDoc).Decode(&result); err != nil {
		res.Err(err)
	}

	return res.Ok(result)
}

func (r *Collection[T]) Upsert(filter *bizdb.BoxFilter, update *bizdb.BoxUpdate, newObj T) *bizresult.Result[struct{}] {
	res := bizresult.New[struct{}]()

	filterDoc := bizdb.FilterToMongo(filter)
	updateDoc := bizdb.UpdateToMongo(update)

	err := r.getCollection().FindOne(r.ctx, filterDoc).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err = r.getCollection().InsertOne(r.ctx, newObj)
			if err != nil {
				return res.Err(err)
			}
		} else {
			return res.Err(err)
		}
	}

	if len(updateDoc) > 0 {
		_, err = r.getCollection().UpdateOne(r.ctx, filterDoc, updateDoc)
		if err != nil {
			return res.Err(err)
		}
	}

	return res.Ok(struct{}{})
}
