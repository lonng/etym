package db

import (
	"etym/pkg/db/model"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func BenchmarkDatabase(b *testing.B) {
	Initialize("root:43OYHrP9Lc2SBeJxwpPA@tcp(172.16.64.135:3306)/etym?charset=utf8")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := model.DbTranslation{Id: int64(i%4000) + 10}
		found, err := database.Cols("id", "word").Get(&record)
		if err != nil || !found {
			b.Fail()
		}
	}
}
