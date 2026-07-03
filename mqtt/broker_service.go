package mqtt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"nav-rain-grid-go/configs"
	"nav-rain-grid-go/domains"
	"strings"
	"sync"
	"time"

	mqttserver "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var BrokerServiceApp = new(BrokerService)

const RainGridSettingsKey = "rain-grid-settings"

type MessageHandler func(clientID string, topic string, payload []byte)

type BrokerService struct {
	mu              sync.RWMutex
	server          *mqttserver.Server
	config          configs.MqttConfig
	startedAt       time.Time
	totalMessages   uint64
	lastMessageAt   int64
	lastClientID    string
	lastTopic       string
	lastPayloadSize int
	handlers        []MessageHandler
}

type BrokerMonitorInfo struct {
	Enable          bool     `json:"enable"`
	Running         bool     `json:"running"`
	Host            string   `json:"host"`
	Port            int      `json:"port"`
	Address         string   `json:"address"`
	HandlerCount    int      `json:"handlerCount"`
	TotalMessages   uint64   `json:"totalMessages"`
	LastMessageAt   int64    `json:"lastMessageAt"`
	LastClientID    string   `json:"lastClientId"`
	LastTopic       string   `json:"lastTopic"`
	LastPayloadSize int      `json:"lastPayloadSize"`
	StartedAt       int64    `json:"startedAt"`
	UptimeSeconds   int64    `json:"uptimeSeconds"`
	Warnings        []string `json:"warnings"`
	CheckedAt       int64    `json:"checkedAt"`
}

type rainGridSettings struct {
	MqttEnable *bool `json:"mqttEnable"`
	MqttPort   int   `json:"mqttPort"`
}

func InitMqtt() {
	RegisterDeviceHeartbeatHandler()
	RegisterPredictHandler()
	if err := BrokerServiceApp.Start(); err != nil {
		panic(err)
	}
}

func (s *BrokerService) Start() error {
	cfg := loadConfig()
	if !cfg.Enable {
		s.mu.Lock()
		s.config = cfg
		s.mu.Unlock()
		zap.L().Info("MQTT服务端未启用")
		return nil
	}

	server, err := newMqttServer(s, cfg)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if s.server != nil {
		s.mu.Unlock()
		_ = server.Close()
		return nil
	}
	s.server = server
	s.config = cfg
	s.startedAt = time.Now()
	s.mu.Unlock()

	go s.serve(server)
	zap.L().Info("MQTT服务端已启动", zap.String("address", cfg.Address()))
	return nil
}

func (s *BrokerService) Stop() {
	s.stopWithConfig(configs.MqttConfig{})
}

func (s *BrokerService) Reload() error {
	cfg := loadConfig()
	if !cfg.Enable {
		s.stopWithConfig(cfg)
		zap.L().Info("MQTT服务端已按配置停用")
		return nil
	}

	s.mu.RLock()
	current := s.server
	currentConfig := s.config
	s.mu.RUnlock()

	if current != nil && mqttConfigEqual(currentConfig, cfg) {
		s.mu.Lock()
		s.config = cfg
		s.mu.Unlock()
		return nil
	}

	server, err := newMqttServer(s, cfg)
	if err != nil {
		return err
	}

	s.mu.Lock()
	oldServer := s.server
	oldConfig := s.config
	if oldServer != nil && mqttConfigEqual(oldConfig, cfg) {
		s.config = cfg
		s.mu.Unlock()
		_ = server.Close()
		return nil
	}
	s.server = server
	s.config = cfg
	s.startedAt = time.Now()
	s.mu.Unlock()

	go s.serve(server)

	if oldServer != nil {
		if err := oldServer.Close(); err != nil {
			zap.L().Error("旧MQTT服务端关闭失败", zap.Error(err))
		}
		zap.L().Info("MQTT服务端已重载",
			zap.String("oldAddress", oldConfig.Address()),
			zap.String("newAddress", cfg.Address()),
		)
		return nil
	}

	zap.L().Info("MQTT服务端已启动", zap.String("address", cfg.Address()))
	return nil
}

func (s *BrokerService) stopWithConfig(cfg configs.MqttConfig) {
	s.mu.Lock()
	server := s.server
	s.server = nil
	s.config = cfg
	s.startedAt = time.Time{}
	s.mu.Unlock()

	if server == nil {
		return
	}
	if err := server.Close(); err != nil {
		zap.L().Error("MQTT服务端关闭失败", zap.Error(err))
	}
	zap.L().Info("MQTT服务端已关闭")
}

func (s *BrokerService) AddMessageHandler(handler MessageHandler) {
	if handler == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, handler)
}

func (s *BrokerService) serve(server *mqttserver.Server) {
	if err := server.Serve(); err != nil {
		zap.L().Error("MQTT服务端运行失败", zap.Error(err))
	}
}

func (s *BrokerService) receive(clientId string, topic string, payload []byte) {
	zap.L().Info("收到MQTT消息",
		zap.String("clientID", clientId),
		zap.String("topic", topic),
		zap.Int("payloadSize", len(payload)),
	)

	s.mu.Lock()
	s.totalMessages++
	s.lastMessageAt = time.Now().UnixMilli()
	s.lastClientID = clientId
	s.lastTopic = topic
	s.lastPayloadSize = len(payload)
	handlers := append([]MessageHandler(nil), s.handlers...)
	s.mu.Unlock()

	for _, handler := range handlers {
		handler(clientId, topic, payload)
	}
}

func (s *BrokerService) Status() BrokerMonitorInfo {
	s.mu.RLock()
	cfg := s.config
	running := s.server != nil
	startedAt := s.startedAt
	info := BrokerMonitorInfo{
		Running:         running,
		HandlerCount:    len(s.handlers),
		TotalMessages:   s.totalMessages,
		LastMessageAt:   s.lastMessageAt,
		LastClientID:    s.lastClientID,
		LastTopic:       s.lastTopic,
		LastPayloadSize: s.lastPayloadSize,
		Warnings:        make([]string, 0),
		CheckedAt:       time.Now().UnixMilli(),
	}
	s.mu.RUnlock()

	if cfg.Port == 0 {
		cfg = loadConfig()
	}
	info.Enable = cfg.Enable
	info.Host = cfg.Host
	info.Port = cfg.Port
	info.Address = cfg.Address()
	if running && !startedAt.IsZero() {
		info.StartedAt = startedAt.UnixMilli()
		info.UptimeSeconds = int64(time.Since(startedAt).Seconds())
	}
	if !info.Enable {
		info.Warnings = append(info.Warnings, "MQTT 服务未启用")
	} else if !info.Running {
		info.Warnings = append(info.Warnings, "MQTT 服务未运行")
	}
	if info.HandlerCount == 0 {
		info.Warnings = append(info.Warnings, "未注册 MQTT 消息处理器")
	}
	return info
}

func loadConfig() configs.MqttConfig {
	yamlCfg := loadYamlConfig()
	if dbCfg, ok := loadDatabaseMqttConfig(yamlCfg); ok {
		return dbCfg
	}
	return yamlCfg
}

func loadYamlConfig() configs.MqttConfig {
	cfg := configs.MqttConfig{
		Enable: true,
		Port:   configs.DefaultMqttPort,
	}
	if global.NAV_VIPER == nil {
		return cfg
	}
	if !global.NAV_VIPER.IsSet("mqtt") {
		return cfg
	}
	if err := global.NAV_VIPER.UnmarshalKey("mqtt", &cfg); err != nil {
		zap.L().Error("解析MQTT配置失败, 使用默认配置", zap.Error(err))
		return configs.MqttConfig{
			Enable: true,
			Port:   configs.DefaultMqttPort,
		}
	}
	if cfg.Port == 0 {
		cfg.Port = configs.DefaultMqttPort
	}
	return cfg
}

func loadDatabaseMqttConfig(defaultConfig configs.MqttConfig) (configs.MqttConfig, bool) {
	if global.NAV_DB == nil {
		return configs.MqttConfig{}, false
	}

	var config domains.Config
	err := global.NAV_DB.Where("`key` = ?", RainGridSettingsKey).Order("id desc").First(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return configs.MqttConfig{}, false
	}
	if err != nil {
		zap.L().Warn("读取数据库MQTT配置失败, 使用YAML配置", zap.Error(err))
		return configs.MqttConfig{}, false
	}
	if strings.TrimSpace(config.Value) == "" {
		return configs.MqttConfig{}, false
	}

	var settings rainGridSettings
	if err := json.Unmarshal([]byte(config.Value), &settings); err != nil {
		zap.L().Warn("解析数据库MQTT配置失败, 使用YAML配置", zap.Error(err))
		return configs.MqttConfig{}, false
	}

	cfg := defaultConfig
	if settings.MqttEnable != nil {
		cfg.Enable = *settings.MqttEnable
	}
	if settings.MqttPort > 0 {
		cfg.Port = settings.MqttPort
	}
	if cfg.Port == 0 {
		cfg.Port = configs.DefaultMqttPort
	}
	return cfg, true
}

func newMqttServer(service *BrokerService, cfg configs.MqttConfig) (*mqttserver.Server, error) {
	server := mqttserver.New(nil)
	if err := server.AddHook(new(authHook), nil); err != nil {
		return nil, fmt.Errorf("添加MQTT认证hook失败: %w", err)
	}
	if err := server.AddHook(new(receiveHook), &receiveHookOptions{service: service}); err != nil {
		return nil, fmt.Errorf("添加MQTT接收hook失败: %w", err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "mqtt-tcp",
		Address: cfg.Address(),
	})
	if err := server.AddListener(tcp); err != nil {
		return nil, fmt.Errorf("监听MQTT端口失败: %w", err)
	}
	return server, nil
}

func mqttConfigEqual(a configs.MqttConfig, b configs.MqttConfig) bool {
	return a.Enable == b.Enable && a.Host == b.Host && a.Port == b.Port
}

type receiveHookOptions struct {
	service *BrokerService
}

type authHook struct {
	mqttserver.HookBase
}

func (h *authHook) ID() string {
	return "nav-rain-mqtt-auth"
}

func (h *authHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqttserver.OnConnectAuthenticate,
		mqttserver.OnACLCheck,
	}, []byte{b})
}

func (h *authHook) OnConnectAuthenticate(cl *mqttserver.Client, pk packets.Packet) bool {
	username := strings.TrimSpace(string(pk.Connect.Username))
	password := strings.TrimSpace(string(pk.Connect.Password))
	if validDeviceCredential(username, password) {
		return true
	}

	clientID := ""
	if cl != nil {
		clientID = cl.ID
	}
	if clientID == "" {
		clientID = pk.Connect.ClientIdentifier
	}
	zap.L().Warn("MQTT客户端认证失败",
		zap.String("clientId", clientID),
		zap.String("username", username),
	)
	return false
}

func (h *authHook) OnACLCheck(cl *mqttserver.Client, topic string, write bool) bool {
	return true
}

func validDeviceCredential(username string, password string) bool {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return false
	}
	return password == username || password == base64.StdEncoding.EncodeToString([]byte(username))
}

type receiveHook struct {
	mqttserver.HookBase
	service *BrokerService
}

func (h *receiveHook) ID() string {
	return "nav-rain-mqtt-receive"
}

func (h *receiveHook) Provides(b byte) bool {
	return bytes.Contains([]byte{mqttserver.OnPublish}, []byte{b})
}

func (h *receiveHook) Init(config any) error {
	options, ok := config.(*receiveHookOptions)
	if !ok || options == nil || options.service == nil {
		return mqttserver.ErrInvalidConfigType
	}
	h.service = options.service
	return nil
}

func (h *receiveHook) OnPublish(cl *mqttserver.Client, pk packets.Packet) (packets.Packet, error) {
	clientId := ""
	if cl != nil {
		clientId = cl.ID
	}
	h.service.receive(clientId, pk.TopicName, pk.Payload)
	return pk, nil
}
