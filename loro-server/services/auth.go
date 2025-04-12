package services

import (
	"context"
	"errors"
	"fmt"
	"server/db"
	"server/db/utils"
	"server/models"
	su "server/utils"
	"time"

	"github.com/jackc/pgx/v5"
)

type AuthService struct {
	pool *db.PostgresPool
}

func NewAuthService(pool *db.PostgresPool) AuthService {
	return AuthService{pool: pool}
}

func (svc AuthService) SignIn(cred models.Credential) (string, error) {
	user := utils.User{}

	err := svc.pool.QueryRow(context.Background(), "select id, username, password from users where username = $1", cred.Username).
		Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err := svc.pool.Execute(context.Background(), "insert into users(username, created_at, password, private_key, public_key) values ($1,$2,$3,$4,$5)",
				cred.Username,
				time.Now(),
				su.CreateHash(cred.Password),
				su.Encrypt(su.GenerateKey(), cred.Password),
				su.GenerateKey())
			if err != nil {
				return "", err
			}
			// user created then token is sent
			return su.MakeToken(*user.Username)
		}
		return "", err
	}

	securePwd := su.CreateHash(cred.Password)
	if *user.Password != securePwd {
		return "", fmt.Errorf("Wrong password")
	}
	return su.MakeToken(*user.Username)
}
