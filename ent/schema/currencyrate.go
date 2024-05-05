package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type CurrencyRate struct {
	ent.Schema
}

func (CurrencyRate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (CurrencyRate) Fields() []ent.Field {
	return []ent.Field{
		field.String("symbol"),
		field.String("rate"),
		field.Time("expired_at").Optional().Nillable(),
	}
}

func (CurrencyRate) Indexes() []ent.Index {
	return []ent.Index{
		// unique index.
		index.Fields("symbol"),
	}
}

// Edges of the User.
func (CurrencyRate) Edges() []ent.Edge {
	return nil
}
