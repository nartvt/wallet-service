package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type Ico struct {
	ent.Schema
}

func (Ico) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (Ico) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.Int32("round_id"),
		field.String("round_name"),
		field.String("price"),
		field.String("num_token"),
		field.Int32("num_sub"),
		field.String("price_gap"), // percent
		field.Time("ended_at").Optional().Nillable(),
	}
}

// Edges of the User.
func (Ico) Edges() []ent.Edge {
	return nil
}
