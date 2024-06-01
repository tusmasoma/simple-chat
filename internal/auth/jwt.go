package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Payload struct {
	JTI    string `json:"jti"`
	UserID string `json:"userId"`
}

const expectedTokenParts = 3

func loadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	// ファイルから秘密鍵をバイトスライスとして読み込む
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading the key file: %w", err)
	}

	// PEMエンコードされたデータからPEMブロックをデコード
	block, _ := pem.Decode(keyBytes)
	if block == nil || (block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY") {
		return nil, fmt.Errorf("failed to decode PEM block containing the key")
	}

	// PEMブロックからRSA秘密鍵をパース
	privInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	privKey, ok := privInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA private key")
	}

	return privKey, nil
}

func loadPublicKeyFromFile(filename string) (*rsa.PublicKey, error) {
	// ファイルから公開鍵をバイトスライスとして読み込む
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading the key file: %w", err)
	}

	// PEMエンコードされたデータからPEMブロックをデコード
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing the key")
	}

	// PEMブロックからRSA公開鍵をパース
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	return pubKey, nil
}

// Base64Urlエンコード
func base64UrlEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// Base64Urlデコード
func base64UrlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// アクセストークン(JWT形式)の生成
func GenerateToken(userID, name string) (string, string) {
	// ヘッダの作成
	header := map[string]string{
		"typ": "JWT",
		"alg": "RS256",
	}
	headerBytes, _ := json.Marshal(header)
	encodedHeader := base64UrlEncode(headerBytes)

	// ペイロードの作成
	jti := uuid.New().String()
	payload := map[string]string{
		"jti":    jti,
		"userId": userID,
		"Name":   name,
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64UrlEncode(payloadBytes)

	// エンコードされたヘッダとペイロードを結合
	jwtWithoutSignature := fmt.Sprintf("%s.%s", encodedHeader, encodedPayload)

	// SHA-256ハッシュを計算
	hashed := sha256.Sum256([]byte(jwtWithoutSignature))

	// 署名作成
	privateKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	privKey, err := loadPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		panic(err)
	}
	encodedSignature := base64UrlEncode(signature)

	// JWTを完成
	jwt := fmt.Sprintf("%s.%s", jwtWithoutSignature, encodedSignature)

	return jwt, jti
}

func ValidateAccessToken(jwt string) error {
	//　アクセストークンの検証
	parts := strings.Split(jwt, ".")
	if len(parts) != expectedTokenParts {
		return fmt.Errorf("invalid token")
	}
	// エンコードされたヘッダとペイロードを結合
	jwtWithoutSignature := fmt.Sprintf("%s.%s", parts[0], parts[1])
	// SHA-256ハッシュを計算
	hashed := sha256.Sum256([]byte(jwtWithoutSignature))

	// 著名作成
	signature, err := base64UrlDecode(parts[2])
	if err != nil {
		return fmt.Errorf("decoding failed: %w", err)
	}

	// 検証
	publicKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	pubKey, err := loadPublicKeyFromFile(publicKeyPath)
	if err != nil {
		return err
	}

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
	log.Print(err)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func GetPayloadFromToken(jwt string) (Payload, error) {
	var emptyPayload Payload
	//　アクセストークンの検証
	parts := strings.Split(jwt, ".")
	if len(parts) != expectedTokenParts {
		return emptyPayload, fmt.Errorf("invalid token")
	}
	// エンコードされたヘッダとペイロードを結合
	encodedPayload := parts[1]
	// Base64Urlデコード
	payloadBytes, err := base64UrlDecode(encodedPayload)
	if err != nil {
		return emptyPayload, fmt.Errorf("decoding failed: %w", err)
	}

	// JSONデコード
	var payload Payload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return emptyPayload, fmt.Errorf("JSON unmarshalling failed")
	}

	return payload, nil
}
