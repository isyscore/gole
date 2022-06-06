package config

import (
	"flag"
	"fmt"
	"github.com/isyscore/gole/util"
	"github.com/isyscore/gole/yaml"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
)

var appProperty *ApplicationProperty
var configExist = false
var loadLock sync.Mutex
var configLoaded = false

// LoadConfig 默认读取./resources/下面的配置文件
// 支持yml、yaml、json、properties格式
// 优先级yaml > yml > properties > json
func LoadConfig() {
	loadLock.Lock()
	defer loadLock.Unlock()
	if configLoaded {
		return
	}

	LoadConfigFromRelativePath("")
	configLoaded = true
}

// LoadConfigFromRelativePath 加载相对文件路径，相对路径是相对系统启动的位置部分
func LoadConfigFromRelativePath(resourceAbsPath string) {
	dir, _ := os.Getwd()
	pkg := strings.Replace(dir, "\\", "/", -1)
	LoadConfigWithAbsPath(path.Join(pkg, "", resourceAbsPath))
}

// LoadConfigWithAbsPath 加载资源文件目录的绝对路径内容，比如：/user/xxx/mmm-biz-service/resources/
// 支持yml、yaml、json、properties格式
// 优先级yaml > yml > properties > json
// 支持命令行：--app.profile xxx
func LoadConfigWithAbsPath(resourceAbsPath string) {
	doLoadConfigFromAbsPath(resourceAbsPath)

	// 读取cm文件
	AppendConfigFromRelativePath("./config/application-default.yml")

	// 加载ApiModule
	ApiModule = GetValueString("api-module")
}

func ExistConfigFile() bool {
	return configExist
}

// AppendConfigFromRelativePath 追加配置：相对路径的配置文件
func AppendConfigFromRelativePath(fileName string) {
	dir, _ := os.Getwd()
	pkg := strings.Replace(dir, "\\", "/", -1)
	fileName = path.Join(pkg, "", fileName)
	extend := getFileExtension(fileName)
	extend = strings.ToLower(extend)
	switch extend {
	case "yaml":
		{
			AppendYamlFile(fileName)
			return
		}
	case "yml":
		{
			AppendYamlFile(fileName)
			return
		}
	case "properties":
		{
			AppendPropertyFile(fileName)
			return
		}
	case "json":
		{
			AppendJsonFile(fileName)
			return
		}
	}
}

// AppendConfigWithAbsPath 追加配置：绝对路径的配置文件
func AppendConfigWithAbsPath(fileName string) {
	extend := getFileExtension(fileName)
	extend = strings.ToLower(extend)
	switch extend {
	case "yaml":
		{
			AppendYamlFile(fileName + fileName)
			return
		}
	case "yml":
		{
			AppendYamlFile(fileName + fileName)
			return
		}
	case "properties":
		{
			AppendPropertyFile(fileName + fileName)
			return
		}
	case "json":
		{
			AppendJsonFile(fileName + fileName)
			return
		}
	}
}

// 多种格式优先级：json > properties > yaml > yml
func doLoadConfigFromAbsPath(resourceAbsPath string) {
	if !strings.HasSuffix(resourceAbsPath, "/") {
		resourceAbsPath += "/"
	}
	files, err := ioutil.ReadDir(resourceAbsPath)
	if err != nil {
		return
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	LoadYamlFile(resourceAbsPath + "application.yaml")
	LoadYamlFile(resourceAbsPath + "application.yml")
	LoadPropertyFile(resourceAbsPath + "application.properties")
	LoadJsonFile(resourceAbsPath + "application.json")

	for _, fileInfo := range files {
		if fileInfo.IsDir() {
			continue
		}

		fileName := fileInfo.Name()
		if !strings.HasPrefix(fileName, "application") {
			continue
		}

		// 默认配置
		if fileName == "application.yaml" {
			configExist = true
			break
		} else if fileName == "application.yml" {
			configExist = true
			break
		} else if fileName == "application.properties" {
			configExist = true
			break
		} else if fileName == "application.json" {
			configExist = true
			break
		}

		profile := getActiveProfile()
		if profile != "" {
			SetValue("base.profiles.active", profile)
			currentProfile := getProfileFromFileName(fileName)
			if currentProfile == profile {
				LoadFile(resourceAbsPath + fileName)
			}
		}
	}
}

// LoadFile 加载某个
func LoadFile(filePath string) {
	extend := getFileExtension(filePath)
	extend = strings.ToLower(extend)
	if extend == "yaml" {
		configExist = true
		LoadYamlFile(filePath)
	} else if extend == "yml" {
		configExist = true
		LoadYamlFile(filePath)
	} else if extend == "properties" {
		configExist = true
		LoadPropertyFile(filePath)
	} else if extend == "json" {
		configExist = true
		LoadJsonFile(filePath)
	}
}

// 临时写死
// 优先级：环境变量 > 本地配置
func getActiveProfile() string {
	var profile string
	flag.StringVar(&profile, "gole.profile", "", "环境变量")
	flag.Parse()
	return profile
}

func GetProperty() *ApplicationProperty {
	return appProperty
}

func getProfileFromFileName(fileName string) string {
	if strings.HasPrefix(fileName, "application-") {
		words := strings.SplitN(fileName, ".", 2)
		appNames := words[0]

		appNameAndProfile := strings.SplitN(appNames, "-", 2)
		return appNameAndProfile[1]
	}
	return ""
}

func getFileExtension(fileName string) string {
	if strings.Contains(fileName, ".") {
		words := strings.SplitN(fileName, ".", 2)
		return words[1]
	}
	return ""
}

func LoadYamlFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	property, err := yaml.YamlToProperties(string(content))
	valueMap, _ := yaml.PropertiesToMap(property)
	appProperty.ValueMap = valueMap

	yamlMap, err := yaml.YamlToMap(string(content))
	appProperty.ValueDeepMap = yamlMap
}

func AppendYamlFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	property, err := yaml.YamlToProperties(string(content))
	valueMap, _ := yaml.PropertiesToMap(property)
	for k, v := range valueMap {
		SetValue(k, v)
	}
}

func LoadPropertyFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	valueMap, _ := yaml.PropertiesToMap(string(content))
	appProperty.ValueMap = valueMap

	yamlStr, _ := yaml.PropertiesToYaml(string(content))
	yamlMap, _ := yaml.YamlToMap(yamlStr)
	appProperty.ValueDeepMap = yamlMap
}

func AppendPropertyFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	valueMap, _ := yaml.PropertiesToMap(string(content))
	for k, v := range valueMap {
		SetValue(k, v)
	}
}

func LoadJsonFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	yamlStr, err := yaml.JsonToYaml(string(content))
	property, err := yaml.YamlToProperties(yamlStr)
	valueMap, _ := yaml.PropertiesToMap(property)
	appProperty.ValueMap = valueMap

	yamlMap, _ := yaml.YamlToMap(yamlStr)
	appProperty.ValueDeepMap = yamlMap
}

func AppendJsonFile(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("fail to read file:", err)
	}

	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}

	yamlStr, err := yaml.JsonToYaml(string(content))
	property, err := yaml.YamlToProperties(yamlStr)
	valueMap, _ := yaml.PropertiesToMap(property)
	for k, v := range valueMap {
		SetValue(k, v)
	}
}

func SetValue(key string, value interface{}) {
	if appProperty == nil {
		appProperty = &ApplicationProperty{}
		appProperty.ValueMap = make(map[string]interface{})
		appProperty.ValueDeepMap = make(map[string]interface{})
	} else if appProperty.ValueMap == nil {
		appProperty.ValueMap = make(map[string]interface{})
	} else if appProperty.ValueDeepMap == nil {
		appProperty.ValueDeepMap = make(map[string]interface{})
	}
	if oldValue, exist := appProperty.ValueMap[key]; exist {
		if !util.IsBaseType(reflect.TypeOf(oldValue)) {
			if reflect.TypeOf(oldValue) != reflect.TypeOf(value) {
				return
			}
		}
	}
	appProperty.ValueMap[key] = value
	doPutValue(key, value)
}

func doPutValue(key string, value interface{}) {
	if strings.Contains(key, ".") {
		oldValue := GetValue(key)
		if nil == oldValue && value != nil {
			if appProperty.ValueDeepMap == nil {
				appProperty.ValueDeepMap = make(map[string]interface{})
			}
			appProperty.ValueDeepMap[key] = value
			return
		}
		if !util.IsBaseType(reflect.TypeOf(oldValue)) {
			if reflect.TypeOf(oldValue).Kind() != reflect.TypeOf(value).Kind() {
				return
			}
		}

		lastIndex := strings.LastIndex(key, ".")
		startKey := key[:lastIndex]
		endKey := key[lastIndex+1:]

		data := GetValue(startKey)
		startValue := util.ToMap(data)
		if nil != startValue {
			startValue[endKey] = value
		}

		doPutValue(startKey, startValue)
	}
	appProperty.ValueDeepMap[key] = value
}

func GetValueString(key string) string {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToString(value)
	}
	return ""
}

func GetValueInt(key string) int {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt(value)
	}
	return 0
}

func GetValueInt8(key string) int8 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt8(value)
	}
	return 0
}

func GetValueInt16(key string) int16 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt16(value)
	}
	return 0
}

func GetValueInt32(key string) int32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt32(value)
	}
	return 0
}

func GetValueInt64(key string) int64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt64(value)
	}
	return 0
}

func GetValueUInt(key string) uint {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt(value)
	}
	return 0
}

func GetValueUInt8(key string) uint8 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt8(value)
	}
	return 0
}

func GetValueUInt16(key string) uint16 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt16(value)
	}
	return 0
}

func GetValueUInt32(key string) uint32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt32(value)
	}
	return 0
}

func GetValueUInt64(key string) uint64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt64(value)
	}
	return 0
}

func GetValueFloat32(key string) float32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToFloat32(value)
	}
	return 0
}

func GetValueFloat64(key string) float64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToFloat64(value)
	}
	return 0
}

func GetValueBool(key string) bool {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToBool(value)
	}
	return false
}

func GetValueStringDefault(key, defaultValue string) string {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToString(value)
	}
	return defaultValue
}

func GetValueIntDefault(key string, defaultValue int) int {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt(value)
	}
	return defaultValue
}

func GetValueInt8Default(key string, defaultValue int8) int8 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt8(value)
	}
	return defaultValue
}

func GetValueInt16Default(key string, defaultValue int16) int16 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt16(value)
	}
	return defaultValue
}

func GetValueInt32Default(key string, defaultValue int32) int32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt32(value)
	}
	return defaultValue
}

func GetValueInt64Default(key string, defaultValue int64) int64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToInt64(value)
	}
	return defaultValue
}

func GetValueUIntDefault(key string, defaultValue uint) uint {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt(value)
	}
	return defaultValue
}

func GetValueUInt8Default(key string, defaultValue uint8) uint8 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt8(value)
	}
	return defaultValue
}

func GetValueUInt16Default(key string, defaultValue uint16) uint16 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt16(value)
	}
	return defaultValue
}

func GetValueUInt32Default(key string, defaultValue uint32) uint32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt32(value)
	}
	return defaultValue
}

func GetValueUInt64Default(key string, defaultValue uint64) uint64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToUInt64(value)
	}
	return defaultValue
}

func GetValueFloat32Default(key string, defaultValue float32) float32 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToFloat32(value)
	}
	return defaultValue
}

func GetValueFloat64Default(key string, defaultValue float64) float64 {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToFloat64(value)
	}
	return defaultValue
}

func GetValueBoolDefault(key string, defaultValue bool) bool {
	if value, exist := appProperty.ValueMap[key]; exist {
		return util.ToBool(value)
	}
	return false
}

func GetValueObject(key string, targetPtrObj interface{}) error {
	data := doGetValue(appProperty.ValueDeepMap, key)
	err := util.DataToObject(data, targetPtrObj)
	if err != nil {
		return err
	}
	return nil
}

func GetValue(key string) interface{} {
	return doGetValue(appProperty.ValueDeepMap, key)
}

func doGetValue(parentValue interface{}, key string) interface{} {
	if key == "" {
		return parentValue
	}
	parentValueKind := reflect.ValueOf(parentValue).Kind()
	if parentValueKind == reflect.Map {
		keys := strings.SplitN(key, ".", 2)
		v1 := reflect.ValueOf(parentValue).MapIndex(reflect.ValueOf(keys[0]))
		emptyValue := reflect.Value{}
		if v1 == emptyValue {
			return nil
		}
		if len(keys) == 1 {
			return doGetValue(v1.Interface(), "")
		} else {
			return doGetValue(v1.Interface(), fmt.Sprintf("%v", keys[1]))
		}
	}
	return nil
}

type ApplicationProperty struct {
	ValueMap     map[string]interface{}
	ValueDeepMap map[string]interface{}
}
