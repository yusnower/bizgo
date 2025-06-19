package bizmgo

import (
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yusnower/gobiz/bizdb"
	"github.com/yusnower/gobiz/bizresult"
)

// FindOptions represents the options for the FindMany operation
type FindOptions[T any] struct {
	collection *Collection[T]
	filter     *bizdb.BoxFilter
	opts       *options.FindOptions
}

// NewFindOptions creates a new FindOptions instance
func NewFindOptions[T any](collection *Collection[T], filter *bizdb.BoxFilter) *FindOptions[T] {
	return &FindOptions[T]{
		collection: collection,
		filter:     filter,
		opts:       options.Find(),
	}
}

// SetLimit sets the maximum number of documents to return
func (f *FindOptions[T]) SetLimit(limit int64) *FindOptions[T] {
	f.opts.SetLimit(limit)
	return f
}

// SetSkip sets the number of documents to skip
func (f *FindOptions[T]) SetSkip(skip int64) *FindOptions[T] {
	f.opts.SetSkip(skip)
	return f
}

// SetSort sets the sort order
func (f *FindOptions[T]) SetSort(sort interface{}) *FindOptions[T] {
	f.opts.SetSort(sort)
	return f
}

// SetProjection sets the projection
func (f *FindOptions[T]) SetProjection(projection interface{}) *FindOptions[T] {
	f.opts.SetProjection(projection)
	return f
}

// SetHint sets the hint
func (f *FindOptions[T]) SetHint(hint interface{}) *FindOptions[T] {
	f.opts.SetHint(hint)
	return f
}

// Execute executes the find operation with the configured options
func (f *FindOptions[T]) Execute() *bizresult.Result[[]T] {
	res := bizresult.New[[]T]()

	var results []T
	filterDoc := bizdb.FilterToMongo(f.filter)

	cursor, err := f.collection.coll.Find(f.collection.ctx, filterDoc, f.opts)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}
	defer cursor.Close(f.collection.ctx)

	err = cursor.All(f.collection.ctx, &results)
	if err != nil {
		return res.SetErr(MongoErr.Err.Wrap(err, filterDoc))
	}

	return res.SetValue(results)
}
