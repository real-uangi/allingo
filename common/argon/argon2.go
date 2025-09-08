/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 09:17
 */

// Package argon

package argon

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"golang.org/x/crypto/argon2"
	"strings"
)

var logger = log.NewStdLogger("common.argon2")

type Params struct {
	Memory     uint32 // 内存占用 (KB)
	Time       uint32 // 迭代次数
	Threads    uint8  // 并发度
	SaltLength uint32 // 盐长度
	KeyLength  uint32 // 哈希长度
	Pepper     []byte
}

var DefaultParams *Params

func init() {
	DefaultParams = &Params{
		Memory:     64 * 1024, // 64MB
		Time:       3,
		Threads:    1,
		SaltLength: 16,
		KeyLength:  32,
	}
	pepperStr := env.Get("CRYPTO_PEPPER")
	if pepperStr != "" {
		DefaultParams.Pepper = []byte(pepperStr)
	} else {
		logger.Warnf("it's recommended to set up CRYPTO_PEPPER")
	}
}

func GenerateFromPassword(password string, p *Params) (string, error) {
	// 生成随机 salt
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 如果设置了pepper，则密码+pepper一起参与哈希
	passWithPepper := []byte(password)
	if len(p.Pepper) > 0 {
		passWithPepper = append(passWithPepper, p.Pepper...)
	}

	hash := argon2.IDKey(passWithPepper, salt, p.Time, p.Memory, p.Threads, p.KeyLength)

	// 将参数、salt 和 hash 编码为字符串存储
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.Memory, p.Time, p.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	return encoded, nil
}

func ComparePassword(password, encodedHash string, pepper []byte) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid encoded hash format: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, fmt.Errorf("unsupported algorithm or version")
	}

	var memory uint32
	var time uint32
	var threads uint8
	var saltBase64, hashBase64 string

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("invalid params segment: %w", err)
	}

	saltBase64 = parts[4]
	hashBase64 = parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false, fmt.Errorf("invalid base64 salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashBase64)
	if err != nil {
		return false, fmt.Errorf("invalid base64 hash: %w", err)
	}

	// 如果设置了pepper，则密码+pepper一起参与哈希
	passWithPepper := []byte(password)
	if len(pepper) > 0 {
		passWithPepper = append(passWithPepper, pepper...)
	}

	hash := argon2.IDKey(passWithPepper, salt, time, memory, threads, uint32(len(expectedHash)))

	if subtleCompare(hash, expectedHash) {
		return true, nil
	}
	return false, nil
}

// 防时序攻击的比较
func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
