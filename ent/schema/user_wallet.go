package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/rs/xid"
)

// User holds the schema definition for the User entity.
type UserWallet struct {
	ent.Schema
}

func (UserWallet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (UserWallet) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").GoType(xid.ID{}).
			DefaultFunc(xid.New).Unique().Immutable(),
		field.String("user_id"),
		field.String("type"), // SYSTEM, USER
		field.String("symbol"),
		// field.Float("balance"),
		field.Bool("is_active"),
		field.String("balance").
			SchemaType(map[string]string{
				dialect.Postgres: "numeric",
			}),
	}
}

func (UserWallet) Indexes() []ent.Index {
	return []ent.Index{
		// unique index.
		index.Fields("user_id", "symbol", "type").
			Unique(),
	}
}

// Edges of the User.
func (UserWallet) Edges() []ent.Edge {
	return nil
}
