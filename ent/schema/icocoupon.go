package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type IcoCoupon struct {
	ent.Schema
}

func (IcoCoupon) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (IcoCoupon) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.String("user_id"),
		field.String("coupon"),
		field.String("reward"),
		field.String("cashback"),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (IcoCoupon) Indexes() []ent.Index {
	return []ent.Index{
		// unique index.
		index.Fields("coupon").Unique(),
	}
}

// Edges of the User.
func (IcoCoupon) Edges() []ent.Edge {
	return nil
}
