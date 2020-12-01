package uller

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/base64"
	"encoding/hex"
)

/*
* 获取字符串32位md5值
* auth guolei at 20191101
* param data
* return string data的32位md5值
*/
func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

/*
* 获取字符串16位md5值
* auth guolei at 20191101
* param data
* return string data的32位md5值
 */
func Get16MD5Encode(data string) string{
	return GetMD5Encode(data)[8:24]
}

/*
* Rc4加密
* auth guolei at 20191101
* param key 加密秘钥
* param data 加密数据
* return string base64后的加密数据
* return err 错误
 */
func Rc4Encrypt(key string,data []byte)(ret []byte,err error){
	keyByte := []byte(key)
	cipher,err := rc4.NewCipher(keyByte)
	if err != nil{
		return ret,err
	}
	dst := data
	cipher.XORKeyStream(dst, data)
	encodeBytes := make([]byte, base64.RawURLEncoding.EncodedLen(len(dst)))
	base64.RawURLEncoding.Encode(encodeBytes,data)
	return encodeBytes,err
}

/*
* Rc4解密
* auth guolei at 20191101
* param key 解密秘钥
* param data base64后的待解密数据
* return byte[] 解密后数据
* return err 错误
 */
func Rc4Decrypt(key string,src []byte)(ret []byte,err error){
	keyByte := []byte(key)
	data := make([]byte, base64.RawURLEncoding.DecodedLen(len(src)))
	base64.RawURLEncoding.Decode(data,src)
	if err != nil{
		return ret,err
	}
	cipher,err := rc4.NewCipher(keyByte)
	if err != nil{
		return ret,err
	}
	dst := data
	cipher.XORKeyStream(dst, data)
	return dst,err
}