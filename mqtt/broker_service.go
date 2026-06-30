package mqtt

import (
	"bytes"
	"fmt"
	"nav-rain-grid-go/configs"
	"sync"
	"time"

	mqttserver "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/wfu-work/nav-common-go-lib/global"
	"go.uber.org/zap"
)

var BrokerServiceApp = new(BrokerService)

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

func InitMqtt() {
	RegisterDeviceHeartbeatHandler()
	RegisterPredictHandler()
	if err := BrokerServiceApp.Start(); err != nil {
		panic(err)
	}
}

func (s *BrokerService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return nil
	}

	cfg := loadConfig()
	if !cfg.Enable {
		zap.L().Info("MQTT服务端未启用")
		return nil
	}

	server := mqttserver.New(nil)
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		return fmt.Errorf("添加MQTT认证hook失败: %w", err)
	}
	if err := server.AddHook(new(receiveHook), &receiveHookOptions{service: s}); err != nil {
		return fmt.Errorf("添加MQTT接收hook失败: %w", err)
	}

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "mqtt-tcp",
		Address: cfg.Address(),
	})
	if err := server.AddListener(tcp); err != nil {
		return fmt.Errorf("监听MQTT端口失败: %w", err)
	}

	s.server = server
	s.config = cfg
	s.startedAt = time.Now()

	go s.serve(server)
	zap.L().Info("MQTT服务端已启动", zap.String("address", cfg.Address()))
	return nil
}

func (s *BrokerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server == nil {
		return
	}
	if err := s.server.Close(); err != nil {
		zap.L().Error("MQTT服务端关闭失败", zap.Error(err))
	}
	s.server = nil
	s.startedAt = time.Time{}
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

type receiveHookOptions struct {
	service *BrokerService
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
