package bizdb

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestUpdateToMongo(t *testing.T) {
	tests := []struct {
		name   string
		update *BoxUpdate
		want   bson.M
	}{
		{
			name:   "Empty update",
			update: Update(),
			want:   bson.M{},
		},
		{
			name:   "Nil update",
			update: nil,
			want:   bson.M{},
		},
		{
			name:   "Set single field",
			update: Update().Set("name", "John"),
			want: bson.M{
				"$set": bson.M{"name": "John"},
			},
		},
		{
			name: "Set multiple fields",
			update: Update().
				Set("name", "John").
				Set("age", 30).
				Set("active", true),
			want: bson.M{
				"$set": bson.M{
					"name":   "John",
					"age":    30,
					"active": true,
				},
			},
		},
		{
			name: "MongoExpr with single operator",
			update: Update().MongoExpr(bson.M{
				"$set": bson.M{"status": "approved"},
			}),
			want: bson.M{
				"$set": bson.M{"status": "approved"},
			},
		},
		{
			name: "MongoExpr with multiple operators",
			update: Update().MongoExpr(bson.M{
				"$set": bson.M{"status": "approved"},
				"$inc": bson.M{"count": 1},
			}),
			want: bson.M{
				"$set": bson.M{"status": "approved"},
				"$inc": bson.M{"count": 1},
			},
		},
		{
			name: "Mix of Set and MongoExpr",
			update: Update().
				Set("name", "John").
				MongoExpr(bson.M{
					"$set": bson.M{"status": "active"},
					"$inc": bson.M{"counter": 1},
				}),
			want: bson.M{
				"$set": bson.M{"name": "John", "status": "active"},
				"$inc": bson.M{"counter": 1},
			},
		},
		{
			name: "Multiple MongoExpr calls",
			update: Update().
				MongoExpr(bson.M{
					"$set": bson.M{"status": "pending"},
					"$inc": bson.M{"counter": 1},
				}).
				MongoExpr(bson.M{
					"$set":  bson.M{"priority": "high"},
					"$push": bson.M{"tags": "urgent"},
				}),
			want: bson.M{
				"$set":  bson.M{"status": "pending", "priority": "high"},
				"$inc":  bson.M{"counter": 1},
				"$push": bson.M{"tags": "urgent"},
			},
		},
		{
			name: "Nested document in Set",
			update: Update().Set("profile", bson.M{
				"firstName": "John",
				"lastName":  "Doe",
				"age":       30,
			}),
			want: bson.M{
				"$set": bson.M{
					"profile": bson.M{
						"firstName": "John",
						"lastName":  "Doe",
						"age":       30,
					},
				},
			},
		},
		{
			name:   "Array in Set",
			update: Update().Set("tags", []string{"important", "urgent"}),
			want: bson.M{
				"$set": bson.M{
					"tags": []string{"important", "urgent"},
				},
			},
		},
		{
			name: "Set with dot notation for nested fields",
			update: Update().
				Set("address.city", "New York").
				Set("address.zip", "10001"),
			want: bson.M{
				"$set": bson.M{
					"address.city": "New York",
					"address.zip":  "10001",
				},
			},
		},
		{
			name: "Complex MongoExpr with nested operators",
			update: Update().MongoExpr(bson.M{
				"$set": bson.M{
					"status": "active",
					"address": bson.M{
						"city": "New York",
						"zip":  "10001",
					},
				},
				"$push": bson.M{
					"history": bson.M{
						"$each": []bson.M{
							{"action": "created", "date": "2023-01-01"},
							{"action": "updated", "date": "2023-01-02"},
						},
					},
				},
			}),
			want: bson.M{
				"$set": bson.M{
					"status": "active",
					"address": bson.M{
						"city": "New York",
						"zip":  "10001",
					},
				},
				"$push": bson.M{
					"history": bson.M{
						"$each": []bson.M{
							{"action": "created", "date": "2023-01-01"},
							{"action": "updated", "date": "2023-01-02"},
						},
					},
				},
			},
		},
		{
			name: "Overwriting fields with Set",
			update: Update().
				Set("status", "pending").
				Set("status", "approved"), // This should overwrite the previous value
			want: bson.M{
				"$set": bson.M{"status": "approved"},
			},
		},
		{
			name: "MongoExpr overriding previous Set",
			update: Update().
				Set("status", "pending").
				MongoExpr(bson.M{
					"$set": bson.M{"status": "approved"},
				}),
			want: bson.M{
				"$set": bson.M{"status": "approved"},
			},
		},
		{
			name: "Set overriding previous MongoExpr",
			update: Update().
				MongoExpr(bson.M{
					"$set": bson.M{"status": "pending"},
				}).
				Set("status", "approved"),
			want: bson.M{
				"$set": bson.M{"status": "approved"},
			},
		},
		{
			name: "Mixed types",
			update: Update().
				Set("name", "John").
				Set("age", 30).
				Set("active", true).
				Set("score", 95.5).
				Set("tags", []string{"a", "b"}).
				Set("metadata", bson.M{"source": "api"}),
			want: bson.M{
				"$set": bson.M{
					"name":     "John",
					"age":      30,
					"active":   true,
					"score":    95.5,
					"tags":     []string{"a", "b"},
					"metadata": bson.M{"source": "api"},
				},
			},
		},
		{
			name: "Merging complex MongoExpr",
			update: Update().MongoExpr(bson.M{
				"$set": bson.M{
					"profile.name": "John",
					"profile.age":  30,
				},
				"$inc": bson.M{
					"counters.visits": 1,
				},
			}).MongoExpr(bson.M{
				"$set": bson.M{
					"profile.status": "active",
				},
				"$inc": bson.M{
					"counters.actions": 1,
				},
			}),
			want: bson.M{
				"$set": bson.M{
					"profile.name":   "John",
					"profile.age":    30,
					"profile.status": "active",
				},
				"$inc": bson.M{
					"counters.visits":  1,
					"counters.actions": 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpdateToMongo(tt.update)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateToMongo() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

// TestUpdateToMongoEdgeCases tests edge cases for UpdateToMongo
func TestUpdateToMongoEdgeCases(t *testing.T) {
	// Test with invalid MongoExpr value
	t.Run("Invalid MongoExpr value", func(t *testing.T) {
		update := Update()
		update.operations = append(update.operations, updateOperation{
			Op:    updateOperator.MongoExpr,
			Value: "not a bson.M", // Invalid type
		})
		got := UpdateToMongo(update)
		want := bson.M{} // Should not add anything for invalid value
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() with invalid MongoExpr = %v, want %v", got, want)
		}
	})

	// Test with empty operations but non-nil update
	t.Run("Empty operations list", func(t *testing.T) {
		update := &BoxUpdate{operations: []updateOperation{}}
		got := UpdateToMongo(update)
		want := bson.M{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() with empty operations = %v, want %v", got, want)
		}
	})

	// Test with complex merging between Set and MongoExpr
	t.Run("Complex merging", func(t *testing.T) {
		update := Update().
			Set("profile.name", "John").
			Set("profile.age", 30).
			MongoExpr(bson.M{
				"$set": bson.M{
					"profile.name":  "Jane", // Should override the previous Set
					"profile.email": "jane@example.com",
				},
			})
		got := UpdateToMongo(update)
		want := bson.M{
			"$set": bson.M{
				"profile.name":  "Jane",
				"profile.age":   30,
				"profile.email": "jane@example.com",
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() with complex merging = %v, want %v", got, want)
		}
	})

	// Test with MongoExpr containing non-bson.M values
	t.Run("MongoExpr with non-bson.M values", func(t *testing.T) {
		update := Update().MongoExpr(bson.M{
			"$set":         bson.M{"status": "active"},
			"$currentDate": true, // Non-bson.M value
		})
		got := UpdateToMongo(update)
		want := bson.M{
			"$set":         bson.M{"status": "active"},
			"$currentDate": true,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() with non-bson.M values = %v, want %v", got, want)
		}
	})

	// Test with empty MongoExpr
	t.Run("Empty MongoExpr", func(t *testing.T) {
		update := Update().MongoExpr(bson.M{})
		got := UpdateToMongo(update)
		want := bson.M{}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() with empty MongoExpr = %v, want %v", got, want)
		}
	})
}

// TestMongoExprSetMerging specifically tests merging of $set operations in MongoExpr
func TestMongoExprSetMerging(t *testing.T) {
	t.Run("Merging $set in multiple MongoExpr calls", func(t *testing.T) {
		update := Update().
			MongoExpr(bson.M{
				"$set": bson.M{
					"name": "John",
					"age":  30,
				},
			}).
			MongoExpr(bson.M{
				"$set": bson.M{
					"name":   "Jane", // Should override
					"email":  "jane@example.com",
					"active": true,
				},
			})

		got := UpdateToMongo(update)
		want := bson.M{
			"$set": bson.M{
				"name":   "Jane",
				"age":    30,
				"email":  "jane@example.com",
				"active": true,
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() $set merging =\n%v\nwant\n%v", got, want)
		}
	})
}

// TestSetWithMongoExprMerging tests how Set operations merge with MongoExpr operations
func TestSetWithMongoExprMerging(t *testing.T) {
	t.Run("Set and MongoExpr $set merging", func(t *testing.T) {
		update := Update().
			Set("name", "John").
			Set("age", 30).
			MongoExpr(bson.M{
				"$set": bson.M{
					"name":   "Jane", // Should override Set("name", "John")
					"email":  "jane@example.com",
					"active": true,
				},
			})

		got := UpdateToMongo(update)
		want := bson.M{
			"$set": bson.M{
				"name":   "Jane",
				"age":    30,
				"email":  "jane@example.com",
				"active": true,
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("UpdateToMongo() Set and MongoExpr merging =\n%v\nwant\n%v", got, want)
		}
	})
}
