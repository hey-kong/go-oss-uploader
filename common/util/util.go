package util

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func PathIsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateDir(dir string) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func RemoveDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}

// GetEnv returns environment variable if present otherwise returns fallback
func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetEnvAsInt returns environment variable as integer if present otherwise returns fallback.
// Error is returned if value can not be converted to integer.
func GetEnvAsInt(key string, fallback int) (int, error) {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("unabled to convert environment variable '%v' to integer", key)
		}
		return i, nil
	}
	return fallback, nil
}

// GetEnvOrPanic returns environment variable if present otherwise panics
func GetEnvOrPanic(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("Missing required environment variable '%v'\n", key))
	}
	return value
}

func GetEnvAsIntOrPanic(key string) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("Missing required environment variable '%v'\n", key))
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("unabled to convert environment variable '%v' to integer", key))
	}
	return i
}

// IsFile checks whether the specified path is a file
func IsFile(path string) bool {
	return !IsDir(path)
}

// IsDir checks whether the given path is a folder
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func ExecShell(name string, arg ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}
