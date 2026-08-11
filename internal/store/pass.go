package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/scrypt"
)

// vaultFile 是加密保险箱在磁盘上的结构。
// salt 在首次创建时随机生成并持久化；所有 secret 共用由主密码 + 该 salt 派生的密钥。
type vaultFile struct {
	Salt    []byte            `json:"salt"`
	Secrets map[string]secret `json:"secrets"` // key 为明文名称，value 为密文
}

// secret 是单条加密记录：ciphertext 为 AES-GCM 密文（含 tag），nonce 为本次加密随机数。
type secret struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
}

const vaultStoreFile = "vault.json"

// scrypt 参数：N=1<<15 在安全性与速度间取得平衡，适合本地个人工具。
const (
	scryptN = 1 << 15
	scryptR = 8
	scryptP = 1
	keyLen  = 32
)

// deriveKey 由主密码与 salt 通过 scrypt 派生出 32 字节 AES-256 密钥。
func deriveKey(password string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, keyLen)
}

// loadVault 读取并解密元数据；文件不存在时返回空 vault（含 nil Salt）。
func loadVault() (vaultFile, error) {
	var vf vaultFile
	notFound, err := loadJSON(vaultStoreFile, &vf)
	if err != nil {
		return vaultFile{}, err
	}
	if notFound {
		return vaultFile{Secrets: map[string]secret{}}, nil
	}
	if vf.Secrets == nil {
		vf.Secrets = map[string]secret{}
	}
	return vf, nil
}

// ensureSalt 保证 vault 有 salt；首次使用时随机生成一个并写回磁盘。
func ensureSalt(vf *vaultFile) error {
	if len(vf.Salt) > 0 {
		return nil
	}
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}
	vf.Salt = salt
	return saveJSON(vaultStoreFile, vf)
}

// encrypt 用派生密钥对明文做 AES-GCM 加密，返回密文与 nonce。
func encrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	// Seal 会把 tag 追加到 ciphertext 末尾。
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// decrypt 用派生密钥解密。
func decrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// SetSecret 将账号口令以“名称 -> 密文”的形式存入加密保险箱。
// password 用于派生主密钥，不会以任何形式落盘。
func SetSecret(name, value, password string) error {
	if name == "" {
		return errors.New("名称不能为空")
	}
	vf, err := loadVault()
	if err != nil {
		return err
	}
	if err := ensureSalt(&vf); err != nil {
		return err
	}
	key, err := deriveKey(password, vf.Salt)
	if err != nil {
		return err
	}
	ct, nonce, err := encrypt(key, []byte(value))
	if err != nil {
		return err
	}
	vf.Secrets[name] = secret{Ciphertext: ct, Nonce: nonce}
	return saveJSON(vaultStoreFile, vf)
}

// GetSecret 用主密码解密取出指定名称的明文口令。
func GetSecret(name, password string) (string, error) {
	vf, err := loadVault()
	if err != nil {
		return "", err
	}
	if len(vf.Salt) == 0 {
		return "", errors.New("保险箱尚未初始化，请先使用 pass set 创建")
	}
	s, ok := vf.Secrets[name]
	if !ok {
		return "", fmt.Errorf("保险箱中没有名为 %q 的记录", name)
	}
	key, err := deriveKey(password, vf.Salt)
	if err != nil {
		return "", err
	}
	plain, err := decrypt(key, s.Ciphertext, s.Nonce)
	if err != nil {
		// 最常见原因：密码错误导致 GCM 校验失败。
		return "", errors.New("解密失败，可能是主密码错误")
	}
	return string(plain), nil
}

// ListSecretNames 返回保险箱中已有的名称列表（不解密内容）。
func ListSecretNames() ([]string, error) {
	vf, err := loadVault()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(vf.Secrets))
	for n := range vf.Secrets {
		names = append(names, n)
	}
	return names, nil
}

// vaultExists 判断保险箱文件是否已存在（供命令层提示用户）。
func vaultExists() bool {
	dir, err := dirPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, vaultStoreFile))
	return err == nil
}
