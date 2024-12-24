package utils

import (
	"crypto/aes"
	"encoding/base64"
	"github.com/titan/tingchao/ccserver/common"
)

// ********************************************************************************************************************
// ECB模式: mysql中AES_DECRYPT函数的实现方式

// AESEncrypt
//
//	mysql中AES_Encrypt函数的实现方式 使用ECB模式
func AESEncrypt(src []byte, key []byte) (encrypted []byte) {
	cipher, _ := aes.NewCipher(generateKey(key))
	length := (len(src) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)

	copy(plain, src)
	pad := byte(len(plain) - len(src))
	for i := len(src); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted = make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(src); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}
	return encrypted
}

// AESDecrypt
//
//	mysql中AES_DECRYPT函数的实现方式 使用ECB模式
func AESDecrypt(encrypted []byte, key []byte) (decrypted []byte) {
	cipher, _ := aes.NewCipher(generateKey(key))
	decrypted = make([]byte, len(encrypted))

	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}
	return decrypted[:trim]
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

func AesEncrypt2Byte(origData []byte) []byte {
	return AESEncrypt(origData, []byte(common.Configs.MessageAesKey))
}

func AesDecrypt2Byte(encrypted []byte) []byte {
	return AESDecrypt(encrypted, []byte(common.Configs.MessageAesKey))
}

// AesEncrypt2Base64Str
//
//	将二进制数组进行AES加密,返回加密后的base64字符串
//	jsonStrByte []byte  待加密的二进制数据
//	@return string	加密后的base64字符串
func AesEncrypt2Base64Str(jsonStrByte []byte) string {
	jsonStrAesByte := AESEncrypt(jsonStrByte, []byte(common.Configs.MessageAesKey))
	jsonStrAesByteBase64 := base64.StdEncoding.EncodeToString(jsonStrAesByte)
	return jsonStrAesByteBase64
}

// AesDecrypt2Base64Str
//
//	将base64字符串进行解码,并AES解密,返回解密后的字符串
//	jsonStr string  AES加密后并base64编码的字符串
//	@return string	base64解码并AES解密后的字符串
func AesDecrypt2Base64Str(jsonStr string) string {
	jsonStrAesByte, _ := base64.StdEncoding.DecodeString(jsonStr)
	jsonStrByte := AESDecrypt(jsonStrAesByte, []byte(common.Configs.MessageAesKey))
	return string(jsonStrByte)
}
