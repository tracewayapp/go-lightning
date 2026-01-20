package repositories

import (
	"database/sql"
	"usercrud/models"

	"github.com/tracewayapp/go-lightning/lit"
)

// if you were to run this with mysql you would need to replace $1, $2... with ?
// this is up to the user to do themselves based on the driver they choose to use
type userRepository struct{}

func (userRepository *userRepository) Create(db *sql.DB, user models.User) (int, error) {
	return lit.Insert(db, &user)
}

func (userRepository *userRepository) FindById(db *sql.DB, id int) (*models.User, error) {
	return lit.SelectSingle[models.User](db, "SELECT id, first_name, last_name, email FROM users WHERE id = $1", id)
}

func (userRepository *userRepository) FindAll(db *sql.DB) ([]*models.User, error) {
	return lit.Select[models.User](db, "SELECT id, first_name, last_name, email FROM users")
}

func (userRepository *userRepository) Update(db *sql.DB, user models.User) error {
	return lit.Update(db, &user, "id = $1", user.Id)
}

func (userRepository *userRepository) Delete(db *sql.DB, id int) error {
	return lit.Delete(db, "DELETE FROM users WHERE id = $1", id)
}

var UserRepository = userRepository{}
