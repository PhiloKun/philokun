package store

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
)

// cryptEnvelope 是文件加解密所用的落盘结构：
// salt 用于 scrypt 派生密钥，nonce 为 AES-GCM 随机数，ciphertext 含 GCM tag。
type cryptEnvelope struct {
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

// EncryptFile 用 password 派生密钥，对 plaintext 做 AES-256-GCM 加密并写入 outPath。
// 每次都会随机生成 salt 与 nonce，因此相同明文多次加密结果不同。
func EncryptFile(plaintext []byte, password, outPath string) error {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	key, err := deriveKey(password, salt)
	if err != nil {
		return err
	}
	ct, nonce, err := encrypt(key, plaintext)
	if err != nil {
		return err
	}
	env := cryptEnvelope{Salt: salt, Nonce: nonce, Ciphertext: ct}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, data, 0o600)
}

// DecryptFile 从 inPath 读取加密信封，用 password 解密出明文。
// 密码错误会导致 GCM 校验失败，返回明确错误。
func DecryptFile(inPath, password string) ([]byte, error) {
	data, err := os.ReadFile(inPath)
	if err != nil {
		return nil, err
	}
	var env cryptEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, errors.New("不是有效的 philokun 加密文件")
	}
	if len(env.Salt) == 0 || len(env.Nonce) == 0 {
		return nil, errors.New("加密文件缺少 salt/nonce")
	}
	key, err := deriveKey(password, env.Salt)
	if err != nil {
		return nil, err
	}
	plain, err := decrypt(key, env.Ciphertext, env.Nonce)
	if err != nil {
		return nil, errors.New("解密失败，可能是密码错误")
	}
	return plain, nil
}
