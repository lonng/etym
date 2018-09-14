package db

import "etym/pkg/db/model"

func syncSchema() {
	database.StoreEngine("InnoDB").Sync2(
		new(model.DbUser),
		new(model.DbHistory),
	)
}
