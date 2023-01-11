package util

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789" +
	"ABCDEGHIJKLMNOPQRSTUVWXYZ"

// InitLogger ...
func InitLogger() *zap.Logger {
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       true,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}
	logger, err := cfg.Build()
	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}

	zap.ReplaceGlobals(logger)

	return logger
}

// StringWithCharset ...
func StringWithCharset(length int, charset string) string {

	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]

	}
	return string(b)
}

// MakeRandText ...
func MakeRandText(length int) string {

	return StringWithCharset(length, charset)
}

// HashPassword returns a bcrypt hash of the input string
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
