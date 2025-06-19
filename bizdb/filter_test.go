package bizdb

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestFilterToMongo(t *testing.T) {
	tests := []struct {
		name   string
		filter *BoxFilter
		want   bson.M
	}{
		{
			name:   "Empty filter",
			filter: Filter(),
			want:   bson.M{},
		},
		{
			name:   "Eq condition",
			filter: Filter().Eq("name", "John"),
			want: bson.M{
				"name": "John",
			},
		},
		{
			name:   "NotEq condition",
			filter: Filter().NotEq("age", 30),
			want: bson.M{
				"age": bson.M{"$ne": 30},
			},
		},
		{
			name:   "Gt condition",
			filter: Filter().Gt("age", 25),
			want: bson.M{
				"age": bson.M{"$gt": 25},
			},
		},
		{
			name:   "Gte condition",
			filter: Filter().Gte("age", 25),
			want: bson.M{
				"age": bson.M{"$gte": 25},
			},
		},
		{
			name:   "Lt condition",
			filter: Filter().Lt("age", 50),
			want: bson.M{
				"age": bson.M{"$lt": 50},
			},
		},
		{
			name:   "Lte condition",
			filter: Filter().Lte("age", 50),
			want: bson.M{
				"age": bson.M{"$lte": 50},
			},
		},
		{
			name:   "In condition",
			filter: Filter().In("status", []interface{}{"active", "pending"}),
			want: bson.M{
				"status": bson.M{"$in": []interface{}{"active", "pending"}},
			},
		},
		{
			name:   "NotIn condition",
			filter: Filter().NotIn("status", []interface{}{"inactive", "deleted"}),
			want: bson.M{
				"status": bson.M{"$nin": []interface{}{"inactive", "deleted"}},
			},
		},
		{
			name:   "Like condition",
			filter: Filter().Like("name", "Jo%"),
			want: bson.M{
				"name": bson.M{"$regex": "Jo.*", "$options": "i"},
			},
		},
		{
			name:   "Complex Like condition",
			filter: Filter().Like("name", "%Smith%"),
			want: bson.M{
				"name": bson.M{"$regex": ".*Smith.*", "$options": "i"},
			},
		},
		{
			name: "Multiple conditions",
			filter: Filter().
				Eq("name", "John").
				Gt("age", 25).
				In("role", []interface{}{"admin", "user"}),
			want: bson.M{
				"name": "John",
				"age":  bson.M{"$gt": 25},
				"role": bson.M{"$in": []interface{}{"admin", "user"}},
			},
		},
		{
			name: "Or condition",
			filter: Filter().Or(
				Filter().Eq("role", "admin"),
				Filter().Eq("role", "manager"),
			),
			want: bson.M{
				"$or": []bson.M{
					{"role": "admin"},
					{"role": "manager"},
				},
			},
		},
		{
			name: "And condition",
			filter: Filter().And(
				Filter().Eq("status", "active"),
				Filter().Gt("age", 25),
			),
			want: bson.M{
				"status": "active",
				"age":    bson.M{"$gt": 25},
			},
		},
		{
			name: "Nested Or and And conditions",
			filter: Filter().Or(
				Filter().And(
					Filter().Eq("status", "active"),
					Filter().Gt("age", 25),
				),
				Filter().And(
					Filter().Eq("status", "pending"),
					Filter().Lte("age", 25),
				),
			),
			want: bson.M{
				"$or": []bson.M{
					{
						"status": "active",
						"age":    bson.M{"$gt": 25},
					},
					{
						"status": "pending",
						"age":    bson.M{"$lte": 25},
					},
				},
			},
		},
		{
			name: "Mixed standard conditions with Or",
			filter: Filter().
				Eq("company", "Acme").
				Or(
					Filter().Eq("role", "admin"),
					Filter().Eq("role", "manager"),
				),
			want: bson.M{
				"company": "Acme",
				"$or": []bson.M{
					{"role": "admin"},
					{"role": "manager"},
				},
			},
		},
		{
			name: "Raw Mongo condition",
			filter: Filter().Mongo(bson.M{
				"createdAt": bson.M{"$exists": true},
			}),
			want: bson.M{
				"createdAt": bson.M{"$exists": true},
			},
		},
		{
			name: "Raw Mongo with $and",
			filter: Filter().Mongo(bson.M{
				"$and": []interface{}{
					bson.M{"priority": "high"},
					bson.M{"deadline": bson.M{"$lt": "2023-12-31"}},
				},
			}),
			want: bson.M{
				"$and": []bson.M{
					{"priority": "high"},
					{"deadline": bson.M{"$lt": "2023-12-31"}},
				},
			},
		},
		{
			name: "Mixed Raw Mongo with standard conditions",
			filter: Filter().
				Eq("status", "active").
				Mongo(bson.M{
					"createdAt": bson.M{"$exists": true},
				}),
			want: bson.M{
				"status":    "active",
				"createdAt": bson.M{"$exists": true},
			},
		},
		{
			name: "Multiple Raw Mongo conditions",
			filter: Filter().
				Mongo(bson.M{"field1": "value1"}).
				Mongo(bson.M{"field2": "value2"}),
			want: bson.M{
				"field1": "value1",
				"field2": "value2",
			},
		},
		{
			name: "Complex mixed query",
			filter: Filter().
				Eq("status", "active").
				Gt("age", 25).
				Mongo(bson.M{"verified": true}).
				Or(
					Filter().Eq("role", "admin"),
					Filter().And(
						Filter().Eq("role", "user"),
						Filter().Gte("level", 5),
					),
					Filter().Mongo(bson.M{"specialAccess": true}),
				),
			want: bson.M{
				"status":   "active",
				"age":      bson.M{"$gt": 25},
				"verified": true,
				"$or": []bson.M{
					{"role": "admin"},
					{
						"role":  "user",
						"level": bson.M{"$gte": 5},
					},
					{"specialAccess": true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterToMongo(tt.filter)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterToMongo() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

// TestFilterToMongoEdgeCases tests edge cases and potential failure scenarios
func TestFilterToMongoEdgeCases(t *testing.T) {
	// Test with nil filter
	t.Run("Nil filter", func(t *testing.T) {
		var filter *BoxFilter = nil
		got := FilterToMongo(filter)
		want := bson.M{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FilterToMongo(nil) = %v, want %v", got, want)
		}
	})

	// Test with an empty Like pattern
	t.Run("Empty Like pattern", func(t *testing.T) {
		filter := Filter().Like("name", "")
		got := FilterToMongo(filter)
		want := bson.M{
			"name": bson.M{"$regex": "", "$options": "i"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FilterToMongo() with empty Like = %v, want %v", got, want)
		}
	})

	// Test with invalid type for Like
	t.Run("Invalid type for Like", func(t *testing.T) {
		filter := Filter()
		filter.conditions = append(filter.conditions, filterItem{
			Field: "name",
			Op:    operator.Like,
			Value: 123, // Not a string
		})
		got := FilterToMongo(filter)
		want := bson.M{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FilterToMongo() with invalid Like type = %v, want %v", got, want)
		}
	})

	// Test with an empty In array
	t.Run("Empty In array", func(t *testing.T) {
		filter := Filter().In("status", []interface{}{})
		got := FilterToMongo(filter)
		want := bson.M{
			"status": bson.M{"$in": []interface{}{}},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FilterToMongo() with empty In array = %v, want %v", got, want)
		}
	})

	// Test with an invalid Mongo filter type
	t.Run("Invalid Mongo filter type", func(t *testing.T) {
		filter := Filter()
		filter.conditions = append(filter.conditions, filterItem{
			Op:    operator.RawMongo,
			Value: "not a bson.M", // Invalid type
		})
		got := FilterToMongo(filter)
		want := bson.M{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("FilterToMongo() with invalid Mongo filter type = %v, want %v", got, want)
		}
	})
}

// TestFilterToMongoPerformance tests the performance of FilterToMongo with a complex filter
func TestFilterToMongoPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a complex filter
	filter := Filter()
	for i := 0; i < 100; i++ {
		subFilter := Filter().
			Eq("field"+string(rune(i)), "value"+string(rune(i))).
			Gt("age"+string(rune(i)), i)
		filter.Or(subFilter)
	}

	// Benchmark the conversion
	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		FilterToMongo(filter)
	}
	elapsed := time.Since(start)
	avgTimeMs := float64(elapsed.Milliseconds()) / float64(iterations)

	t.Logf("Average time for complex FilterToMongo: %.2f ms", avgTimeMs)
}
