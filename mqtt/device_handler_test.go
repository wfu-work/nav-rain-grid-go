package mqtt

import "testing"

func TestParseDeviceHeartbeatWithAliasAndLocation(t *testing.T) {
	payload := []byte(`{"sncode":"DEV001","alias":"一号设备","lat":30.52,"lng":114.31}`)

	heartbeat := ParseDeviceHeartbeat("device/heartbeat", payload)

	if heartbeat.Sncode != "DEV001" {
		t.Fatalf("unexpected sncode: %s", heartbeat.Sncode)
	}
	if heartbeat.Alias != "一号设备" {
		t.Fatalf("unexpected alias: %s", heartbeat.Alias)
	}
	if heartbeat.Lat == nil || *heartbeat.Lat != 30.52 {
		t.Fatalf("unexpected lat: %#v", heartbeat.Lat)
	}
	if heartbeat.Lng == nil || *heartbeat.Lng != 114.31 {
		t.Fatalf("unexpected lng: %#v", heartbeat.Lng)
	}
}

func TestParseDeviceHeartbeatNestedPayload(t *testing.T) {
	payload := []byte(`{"data":{"snCode":"DEV002","deviceName":"二号设备","latitude":"30.53","longitude":"114.32"}}`)

	heartbeat := ParseDeviceHeartbeat("device/heartbeat", payload)

	if heartbeat.Sncode != "DEV002" {
		t.Fatalf("unexpected sncode: %s", heartbeat.Sncode)
	}
	if heartbeat.Alias != "二号设备" {
		t.Fatalf("unexpected alias: %s", heartbeat.Alias)
	}
	if heartbeat.Lat == nil || *heartbeat.Lat != 30.53 {
		t.Fatalf("unexpected lat: %#v", heartbeat.Lat)
	}
	if heartbeat.Lng == nil || *heartbeat.Lng != 114.32 {
		t.Fatalf("unexpected lng: %#v", heartbeat.Lng)
	}
}

func TestParseDeviceHeartbeatTextFallback(t *testing.T) {
	heartbeat := ParseDeviceHeartbeat("device/heartbeat", []byte("DEV003"))

	if heartbeat.Sncode != "DEV003" {
		t.Fatalf("unexpected sncode: %s", heartbeat.Sncode)
	}
	if heartbeat.Alias != "" || heartbeat.Lat != nil || heartbeat.Lng != nil {
		t.Fatalf("unexpected extra fields: %#v", heartbeat)
	}
}
