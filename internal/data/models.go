package data

import (
	"database/sql"
	"errors"
	"github.com/e-inwork-com/golang-user-microservice/pkg/data"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users data.UserModel
}

func InitModels(db *sql.DB) Models {
	return Models{
		Users: data.UserModel{DB: db},
	}
}
