package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type IcoHistory struct {
	ent.Schema
}

func (IcoHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (IcoHistory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.Int32("round_id"),
		field.String("user_id"),
		field.String("price"),
		field.String("num_token"),
		field.Int32("sub_round"),
		field.String("type").Optional(),
	}
}

// Edges of the User.
func (IcoHistory) Edges() []ent.Edge {
	return nil
}
