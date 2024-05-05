package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type IcoRound struct {
	ent.Schema
}

func (IcoRound) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (IcoRound) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.Int32("round_id"),
		field.Int32("sub_round"),
		field.String("price"),
		field.String("num_token"),
		field.String("bought_token"),
		field.Time("start_at").Optional().Nillable(),
		field.Time("end_at").Optional().Nillable(),
		field.Bool("is_close").Default(false),
	}
}

// Edges of the User.
func (IcoRound) Edges() []ent.Edge {
	return nil
}
