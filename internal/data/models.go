package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users UserModelInterface
}

func InitModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
