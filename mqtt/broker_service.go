package mqtt

import (
	"bytes"
	"fmt"
	"nav-rain-grid-go/configs"
	"sync"

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
	mu       sync.RWMutex
	server   *mqttserver.Server
	config   configs.MqttConfig
	handlers []MessageHandler
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

	s.mu.RLock()
	handlers := append([]MessageHandler(nil), s.handlers...)
	s.mu.RUnlock()

	for _, handler := range handlers {
		handler(clientId, topic, payload)
	}
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
