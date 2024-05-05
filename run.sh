
go run ./cmd/server -conf configs/config.yaml


go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/modifier,sql/upsert,sql/execquery,sql/versioned-migration ./ent/schema