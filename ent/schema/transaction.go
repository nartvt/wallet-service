package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type Transaction struct {
	ent.Schema
}

func (Transaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.String("trans_type"), // ICO, SUBSCRIPTION, FEE, DEPOSITE
		field.String("source"),
		field.String("src_symbol"),
		field.String("src_amount"),
		field.String("destination"),
		field.String("dest_symbol"),
		field.String("dest_amount"),
		field.String("rate"),
		field.String("source_service"), // which from service
		field.String("source_id"),      // which from service
		field.String("status"),
	}
}

// Edges of the User.
func (Transaction) Edges() []ent.Edge {
	return nil
}
