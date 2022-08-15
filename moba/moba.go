package moba

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"gopkg.in/ini.v1"
)

const (
	CRYPTPROTECT_UI_FORBIDDEN = 0x1
)

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procEncryptData = dllcrypt32.NewProc("CryptProtectData")
	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")

	keyHead = "AQAAANCMnd8BFdERjHoAwE/Cl+s="
)

type Moba struct {
	key []byte
	iv  []byte
	cfg *ini.File
}

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewMoba(name string) (*Moba, error) {

	cfg, err := ini.LoadSources(ini.LoadOptions{
		SkipUnrecognizableLines: true,
		KeyValueDelimiters:      "=",
	}, name)

	if err != nil {
		return nil, err
	}

	sessP := []byte(cfg.Section("Misc").Key("SessionP").String())
	LastUsername := cfg.Section("Sesspass").Key("LastUsername").String()
	LastComputername := cfg.Section("Sesspass").Key("LastComputername").String()
	sesspass := cfg.Section("Sesspass").Key(LastUsername + "@" + LastComputername).String()

	key := append(base64Decode(keyHead), base64Decode(sesspass)...)

	var outblob DATA_BLOB
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(NewBlob(key))), 0, uintptr(unsafe.Pointer(NewBlob(sessP))), 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))

	dkey := base64Decode(outblob.ToString())

	return &Moba{
		key: dkey[0:32],
		iv:  aesEncryptECB(dkey[0:32]),
		cfg: cfg,
	}, nil
}

func (m *Moba) Decrypt(in string) string {
	return string(aesDecryptCFB(base64Decode(in), m.key, m.iv))
}

func (m *Moba) ShowPasswords() {
	for _, v := range m.cfg.Section("Passwords").Keys() {
		fmt.Printf("%s: %s \n", v.Name(), m.Decrypt(v.String()))
	}
}

func (m *Moba) ShowCredentials() {
	for _, v := range m.cfg.Section("Credentials").Keys() {
		data := strings.Split(v.String(), ":")
		if len(data) == 2 {
			fmt.Printf("%s: %s %s \n", v.Name(), data[0], m.Decrypt(data[1]))
		}
	}
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToString() string {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return string(d)
}

func base64Decode(str string) []byte {

	buf, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return []byte{}
	}
	return buf
}

func aesEncryptECB(key []byte) []byte {
	cipher, _ := aes.NewCipher(key)
	plain := make([]byte, aes.BlockSize)
	for i := range plain {
		plain[i] = '\x00'
	}
	encrypted := make([]byte, len(plain))
	cipher.Encrypt(encrypted, plain)
	return encrypted
}

func aesDecryptCFB(data, key, iv []byte) []byte {
	block, _ := aes.NewCipher(key)
	stream := NewDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)
	return decrypted
}
