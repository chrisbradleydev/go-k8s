package utils

import (
	"fmt"
	"os"
)

func GetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("failed to get home directory")
	}
	return homeDir
}