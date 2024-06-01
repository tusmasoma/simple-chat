package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/tusmasoma/simple-chat/internal/auth"
	"github.com/tusmasoma/simple-chat/repository"
)

type UserUseCase interface {
	//CreateUserAndGenerateToken(ctx context.Context, email string, passward string) (string, error)
	LoginAndGenerateToken(ctx context.Context, email string, passward string) (string, error)
	//LogoutUser(ctx context.Context, userID string) error
}

type userUseCase struct {
	ur  repository.UserRepository
	ucr repository.UserCacheRepository
}

func NewUserUseCase(ur repository.UserRepository, ucr repository.UserCacheRepository) UserUseCase {
	return &userUseCase{
		ur:  ur,
		ucr: ucr,
	}
}

func (uuc *userUseCase) LoginAndGenerateToken(ctx context.Context, name string, passward string) (string, error) {
	user, err := uuc.ur.GetByName(ctx, name)
	if err != nil {
		log.Printf("Error retrieving user by email")
		return "", err
	}
	// 既にログイン済みかどうか確認する
	session, _ := uuc.ucr.GetUserSession(ctx, user.ID)
	if session != "" {
		log.Printf("Already logged in")
		return "", fmt.Errorf("user id in cache")
	}

	// Clientから送られてきたpasswordをハッシュ化したものとMySQLから返されたハッシュ化されたpasswordを比較する
	if err = auth.CompareHashAndPassword(user.Password, passward); err != nil {
		log.Printf("password does not match")
		return "", err
	}

	jwt, jti := auth.GenerateToken(user.ID, user.Name)
	if err = uuc.ucr.SetUserSession(ctx, user.ID, jti); err != nil {
		log.Print("Failed to set access token in cache")
		return "", err
	}
	return jwt, nil
}
