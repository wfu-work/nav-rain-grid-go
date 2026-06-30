package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type AppConfig struct {
	MqttConfig MqttConfig `mapstructure:"mqtt" json:"mqtt" yaml:"mqtt"`
}

var App = AppConfig{
	MqttConfig: MqttConfig{
		Enable: true,
		Port:   DefaultMqttPort,
	},
}

func NewAppConfig() AppConfig {
	v := viper.New()
	config := GetConfigPath()
	_, err := os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		exePath, _ := os.Executable()
		_ = os.Chdir(filepath.Dir(exePath))
		fmt.Printf("默认配置文件路径不存在, 切换程序目录: %s\n", filepath.Dir(exePath))
	}
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err = v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	appConfig := &AppConfig{}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("config file changed: %v\n", e)
		if err = v.Unmarshal(appConfig); err != nil {
			fmt.Println(err)
			return
		}
	})
	if err = v.Unmarshal(appConfig); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}
	return *appConfig
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() (config string) {
	switch runtime.GOOS {
	case "windows":
		config = "config-win.yaml"
	default:
		config = "config.yaml"
	}
	fmt.Printf("您正在使用 gin 的 %s 模式运行, config 的路径为 %s\n", gin.Mode(), config)

	_, err := os.Stat(config)
	if err != nil || os.IsNotExist(err) {
		config = "config.yaml"
		fmt.Printf("配置文件路径不存在, 使用默认配置文件路径: %s\n", config)
	}
	return config
}
