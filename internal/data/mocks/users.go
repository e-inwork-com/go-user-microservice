package mocks

import (
	"time"

	"github.com/e-inwork-com/go-user-service/internal/data"
	"github.com/google/uuid"
)

type UserModel struct{}

func (m UserModel) Insert(user *data.User) error {
	user.ID = MockFirstUUID()
	user.CreatedAt = time.Now()
	user.Version = 1

	return nil
}

func (m UserModel) GetByID(id uuid.UUID) (*data.User, error) {
	if MockFirstUUID() == id {
		var user = &data.User{
			ID:        id,
			CreatedAt: time.Now(),
			Email:     "jon@doe.com",
			FirstName: "Jon",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}
		return user, nil
	}

	if MockSecondUUID() == id {
		var user = &data.User{
			ID:        id,
			CreatedAt: time.Now(),
			Email:     "nina@doe.com",
			FirstName: "nina",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}
		return user, nil
	}

	return nil, data.ErrRecordNotFound
}

func (m UserModel) GetByEmail(email string) (*data.User, error) {
	if email == "jon@doe.com" {
		var user = &data.User{
			ID:        MockFirstUUID(),
			CreatedAt: time.Now(),
			Email:     "jon@doe.com",
			FirstName: "Jon",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}
		user.Password.Set("pa55word")

		return user, nil
	}

	if email == "nina@doe.com" {
		var user = &data.User{
			ID:        MockSecondUUID(),
			CreatedAt: time.Now(),
			Email:     "nina@doe.com",
			FirstName: "Nina",
			LastName:  "Doe",
			Activated: true,
			Version:   1,
		}
		user.Password.Set("pa55word")

		return user, nil
	}

	return nil, data.ErrRecordNotFound
}

func (m UserModel) Update(user *data.User) error {
	user.Version += 1

	return nil
}
