package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GetToken loads .env if not already loaded and returns the token for the given sale source.
func GetToken(saleSource string) (string, error) {
	_ = godotenv.Load() // Safe to call multiple times; only loads once
	switch strings.ToLower(saleSource) {
	case "21":
		return os.Getenv("C21_API_TOKEN"), nil
	case "asc":
		return os.Getenv("ASC_API_TOKEN"), nil
	case "bjp":
		return os.Getenv("BJP_API_TOKEN"), nil
	case "bsc":
		return os.Getenv("BSC_API_TOKEN"), nil
	case "gtg":
		return os.Getenv("GTG_API_TOKEN"), nil
	case "oat":
		return os.Getenv("OAT_API_TOKEN"), nil
	case "sm":
		return os.Getenv("SMD_API_TOKEN"), nil
	default:
		return "", fmt.Errorf("invalid sale source: must be '21', 'asc', 'bjp', 'bsc', 'gtg', 'oat', or 'sm'")
	}
}
