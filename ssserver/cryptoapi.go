package ssserver

import (
	"crypto/rand"
)

type AEAD interface {
	// NonceSize返回必须传递给Seal的随机数的大小
	// and Open.
	NonceSize() int

	// Overhead返回以下两者间的最大差异
	// plaintext 和 its ciphertext.
	Overhead() int

	// 密封加密和验证明文，验证
	// 附加数据并将结果附加到dst，并返回更新
	// slice。 nonce必须是NonceSize()字节长且对所有人都是唯一的
	// time, 对于给定的密钥。
	//
	// 明文和dst可能完全或根本不是别名。 重用
	// 明文的加密输出存储，使用 plaintext[:0]作为dst。
	Seal(dst, nonce, plaintext, additionalData []byte) []byte

	// 打开解密并验证密文，验证密文
	// 额外的数据，如果成功的话，附加结果明文
	// 到dst，返回更新的片。 nonce必须是NonceSize()
	// 字节长，它和附加数据必须匹配
	// 值传递给Seal。
	//
	// 密文和dst可以完全混淆或根本不混淆。 重用
	// 密文的解密输出存储，使用ciphertext [:0]作为dst。
	//
	// 即使该功能失败，dst的内容，直到其容量，
	// 可能会被覆盖。
	Open(dst, nonce, ciphertext, additionalData []byte) ([]byte, error)
}

// Generates random bytes
func RandomKey(bytes int) ([]byte, error) {
	b := make([]byte, bytes)
	_, err := rand.Read(b)
	return b, err
}

// Encrypts the message in place with the nonce and key and optional additional buffer
func Encrypt(aead AEAD, dst []byte, nonce, plaintext, additionalData []byte) error {
	aead.Seal(dst[:0], nonce, plaintext, additionalData)
	return nil
}

// Decrypts the message with the nonce and key and optional additional buffer returning a copy
// byte slice
func Decrypt(aead AEAD, dst []byte, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	return aead.Open(dst[:0], nonce, ciphertext, additionalData)
}
