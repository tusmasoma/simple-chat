package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/tusmasoma/simple-chat/usecase"
)

type UserHandler interface {
	//CreateUser(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	//Logout(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	uur usecase.UserUseCase
}

func NewUserHandler(uur usecase.UserUseCase) UserHandler {
	return &userHandler{
		uur: uur,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (uh *userHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var requestBody LoginRequest
	if ok := isValidLoginRequest(r.Body, &requestBody); !ok {
		http.Error(w, "Invalid user create request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	jwt, err := uh.uur.LoginAndGenerateToken(ctx, requestBody.Email, requestBody.Password)
	if err != nil {
		http.Error(w, "Failed to Login or generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+jwt)
	w.WriteHeader(http.StatusOK)
}

func isValidLoginRequest(body io.ReadCloser, requestBody *LoginRequest) bool {
	// リクエストボディのJSONを構造体にデコード
	if err := json.NewDecoder(body).Decode(requestBody); err != nil {
		log.Printf("Invalid request body: %v", err)
		return false
	}
	if requestBody.Email == "" || requestBody.Password == "" {
		log.Printf("Missing required fields: Name or Password")
		return false
	}
	return true
}
