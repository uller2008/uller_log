/*
工具类
auth by guolei
*/

package uller

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/mojocn/base64Captcha"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

/*
* 获取当前运行程序的运行路径
* auth guolei at 20191128
* return 程序文件运行路径
 */
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.Replace(dir, "\\", "/", -1)
}

/*
* 判断文件是否存在
* auth guolei at 20191128
* param file 文件全路径
* return true文件存在，false不存在
 */
func FileExists(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

/*
* 获取中文字符在字符串中的位置
* auth guolei at 20191101
* param str 要检索的字符串
* param substr 包含的字符串
* return int 包含的字符串出现的位置，-1未检索到
 */
func UnicodeIndex(str,substr string) int {
	result := strings.Index(str,substr)
	if result != -1 {
		prefix := []byte(str)[0:result]
		rs := []rune(string(prefix))
		result = len(rs)
	}
	return result
}

/*
* http 协程请求
* auth guolei at 20191101
* param url 请求url全路径带http、https
* param method 请求方式get、post...
* param ch 通道，请求结果通知给chan
* param header 自定义请求头
* param data 发送请求数据
 */
func HttpRequest(url string,method string,ch chan []byte,header map[string]string,data string)(){
	req, err := http.NewRequest(strings.ToUpper(method),url,strings.NewReader(data))
	for key,value := range header {
		req.Header.Set(key, value)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println(err)
	}else{
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println(string(body))
		ch <- body
	}
}

/*
* 随机生成数字、字母字符串
* auth guolei at 20191101
* param l 生成字符串长度
* return string 随机生成字符串
 */
func  GetRandomString(length int) string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return strings.ToLower(string(result))
}

/*
* 邮箱验证
* auth guolei at 20191101
* param email 邮箱
* return bool true验证通过，false验证不通过
 */
func VerifyEmailFormat(email string) bool {
	//pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

/*
* 手机号验证
* auth guolei at 20191101
* param mobile 手机号
* return bool true验证通过，false验证不通过
 */
func VerifyMobileFormat(mobile string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobile)
}

/*
* 数字数组转换为英文逗号分隔的字符串
* auth guolei at 20191105
* param obj 数字数组
* return 以逗号分隔的字符串
 */
func IntArrToString(obj []int)(ret string){
	for i := 0; i < len(obj); i++ {
		ret += strconv.Itoa((obj[i])) + ","
	}
	return ret[0:len(ret)-1]
}

/*
* 字符串数组转换为英文逗号分隔的字符串
* auth guolei at 20191105
* param obj 字符串数组
* return 以逗号分隔的字符串
 */
func StringArrToString(obj []string)(ret string){
	for i := 0; i < len(obj); i++ {
		ret += "'" + obj[i] + "',"
	}
	return ret[0:len(ret)-1]
}

/*
* 验证字符串是否为数字
* auth guolei at 20191105
* param string 字符串
* return bool
 */
func IsNum(s string) bool {
	_, err := strconv.ParseInt(s,0,64)
	return err == nil
}

func OffsetXYAfterRotationCore(W, H, L, T, Angle float64) (x, y float64) {

	var DX, DY, X, Y float64

	AngleRad := Angle * math.Pi / 180
	SinX := math.Sin(AngleRad)
	CosX := math.Cos(AngleRad)

	//0<=Angle <=90
	if Angle >= 0 && Angle <= 90 {
		DX = SinX * H
		DY = 0
		X = L - DX
		Y = T - DY
		//fmt.Println("At last Angle,X,Y,DX,DY=", Angle, X, Y, DX, DY)
	} else if Angle > 90 && Angle <= 180 {
		//90<=Angle <=180
		//SinX2 := math.Sin((180 - Angle) )
		//CosX2 := math.Cos((180 - Angle) )
		SinX2 := SinX
		CosX2 := -CosX
		DX = SinX2*H + W*CosX2
		DY = H * CosX2
		X = L - DX
		Y = T - DY
		//fmt.Println("At last Angle,X,Y,DX,DY=", Angle, X, Y, DX, DY)
	} else if Angle > 180 && Angle <= 270 {
		//SinX2 := math.Sin((270 - Angle))
		//CosX2 := math.Cos((270 - Angle))
		SinX2 := -CosX
		CosX2 := -SinX
		DX = SinX2 * W
		DY = CosX2*W + SinX2*H
		X = L - DX
		Y = T - DY
		//fmt.Println("At last Angle,X,Y,DX,DY=", Angle, X, Y, DX, DY)
	} else {
		//SinX2 := math.Sin((360 - Angle))
		SinX2 := -SinX

		DX = 0
		DY = SinX2 * W
		X = L - DX
		Y = T - DY
		//fmt.Println("At last Angle,X,Y,DX,DY=", Angle, X, Y, DX, DY)
	}

	x = X
	y = Y

	return
}

/*
* 验证码
* auth guolei at 20191105
* param mode 输出验证码类型：audio声音验证码，character字符串验证码，digit数学计算验证码
* param len 验证码长度
* return bool
 */
func Captcha(mode string,len int)(ret map[string]string){
	if mode == ""{
		mode = "digit "
	}
	var config interface{}
	switch mode {
	case "audio":
		//声音验证码配置
		var configAudio = base64Captcha.ConfigAudio{
			CaptchaLen: len,
			Language:   "zh",
		}
		config = configAudio
	case "character":
		var configCharacter = base64Captcha.ConfigCharacter{
			Height:             60,
			Width:              240,
			Mode:               base64Captcha.CaptchaModeNumber,
			ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
			ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
			IsUseSimpleFont:    true,
			IsShowHollowLine:   false,
			IsShowNoiseDot:     true,
			IsShowNoiseText:    false,
			IsShowSlimeLine:    false,
			IsShowSineLine:     false,
			CaptchaLen:         len,
		}
		config = configCharacter
	default:
		var configCharacter = base64Captcha.ConfigCharacter{
			Height:             60,
			Width:              240,
			Mode:               base64Captcha.CaptchaModeArithmetic,
			ComplexOfNoiseText: base64Captcha.CaptchaComplexLower,
			ComplexOfNoiseDot:  base64Captcha.CaptchaComplexLower,
			IsUseSimpleFont:    true,
			IsShowHollowLine:   false,
			IsShowNoiseDot:     true,
			IsShowNoiseText:    false,
			IsShowSlimeLine:    false,
			IsShowSineLine:     false,
			CaptchaLen:         len,
		}
		config = configCharacter
		config = configCharacter
	}

	captchaId, captcaInterfaceInstance := base64Captcha.GenerateCaptcha("", config)
	base64blob := base64Captcha.CaptchaWriteToBase64Encoding(captcaInterfaceInstance)
	ret = make(map[string]string)
	ret["captchaId"] = captchaId
	ret["captcha"] = base64blob
	return
}

func VerifyCaptcha(captchaId string, captcha string) bool {
	return base64Captcha.VerifyCaptcha(captchaId, captcha)
}

/*
* 获取雪花算法id
* auth guolei at 20200429
* return id
 */
func SnowflakeId() (id int64) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	id = node.Generate().Int64()
	return id
}

/*
* []map转逗号分隔字符串
* auth guolei at 20200429
* return s
 */
func MapToSplitString(m []map[string]interface{},field string) (s string) {
	for i:=0;i< len(m);i++{
		s += m[i][field].(string) + ","
	}
	s = s[0:len(s)-1]
	return s
}

/*
* Strval 数据类型转换位string，浮点型 3.0将会转换成字符串3, "3"，非数值或字符类型的变量将会被转换成JSON格式字符串
* auth guolei at 20200429
*
 */
func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		fmt.Println(value)
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

/*
* IntRange 随机数生成
* auth guolei at 20200821
* param min 随机数最小值
* param max 随机数最大值
* return int 随机数
 */
func IntRange(min, max int) (int, error) {
	var result int
	switch {
	case min > max:
		// Fail with error
		return result, errors.New("Min cannot be greater than max.")
	case max == min:
		result = max
	case max > min:
		maxRand := max - min

		b, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return result, err
		}
		result = min + int(b.Int64())
	}
	return result, nil
}

/*
* 字符串数组搜索
* auth guolei at 20200902
* param arr 要检索的数组
* param target 检索的字符串
* return int 匹配到的数组索引值
 */
func StringArrSearch(arr []string, target string)(ret int){
	ret = -1
	for i:=0;i<len(arr);i++{
		if arr[i] == target{
			ret = i
			break
		}
	}
	return ret
}

/*
* 按字母顺序排序map
* auth guolei at 20200902
* param params 要排序的map
* return ret 完成排序的map
 */
func MapInStringOrder(params map[string]string)(ret map[string]string) {
	ret = make(map[string]string)
	keys := make([]string, 0)
	for k, _ := range(params) {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ret[k] = params[k]
	}
	return
}

/*
* 仿php http_build_query，将数组数据修改成post的以&符区隔参数
* auth guolei at 20200902
* param params 要排序的map
* return ret 完成排序的map
 */
func HttpBuildQuery(params map[string]string)(ret string) {
	if len(params) == 0{
		return
	}
	for k,v := range(params) {
		ret += k + "=" + v + "&"
	}
	ret = ret[0:len(ret) - 1]
	return
}

/*
* 获取本机ip，如果本机有多个ip地址则返回第一个
* auth guolei at 20200902
* return ip地址
 */
func GetLocalIP()(ip string){
	ip = ""
	addrs ,err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}
	return
}

func BytesToUint32(array []byte) uint32 {
	var data uint32 =0
	for i:=0;i< len(array);i++  {
		data = data+uint32(uint(array[i])<<uint(8*i))
	}

	return data
}

func IsLittleEndian() bool {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	return (b == 0x04)
}

func IntToBytes(num int)[4]byte{
	x := int32(num)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	var ret [4]byte
	for i:=0;i<len(ret);i++{
		ret[i] = bytesBuffer.Bytes()[i]
	}
	return ret
}

//int64转[]byte
func Int64ToBytes(num int64)[8]byte{
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, num)
	var ret [8]byte
	for i:=0;i<len(ret);i++{
		ret[i] = bytesBuffer.Bytes()[i]
	}
	return ret
}

//字节转换成整形
func BytesToInt(b [4]byte) int {
	bytesBuffer := bytes.NewBuffer(b[0:4])
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

//字节转换成int64
func BytesToInt64(b [8]byte) int64 {
	return int64(binary.BigEndian.Uint64(b[0:8]))
}

//[]byte合并
func BytesCombine(pBytes ...[]byte) []byte {
	return bytes.Join(pBytes, []byte(""))
}

//struct数组转为map数组
func StructArrToMap(obj []interface{}) []map[string]interface{} {
	var data = []map[string]interface{}{}
	for k,v := range obj{
		data[k] = StructToMap(v)
	}
	return data
}

//struct转换为map
func StructToMap(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}