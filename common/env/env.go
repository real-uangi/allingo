/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 11:20
 */

// Package env

package env

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file", err)
	}
}

func GetOrDefault(key string, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Get(key string) string {
	return os.Getenv(key)
}

func GetInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	i, _ := strconv.Atoi(v)
	return i
}

func GetIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
