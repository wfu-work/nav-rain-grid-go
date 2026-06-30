package mqtt

import (
	"encoding/base64"
	"testing"
)

func TestValidDeviceCredential(t *testing.T) {
	sncode := "RAIN-001"
	encoded := base64.StdEncoding.EncodeToString([]byte(sncode))

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{name: "plain sncode password", username: sncode, password: sncode, want: true},
		{name: "base64 sncode password", username: sncode, password: encoded, want: true},
		{name: "trim spaces", username: " " + sncode + " ", password: " " + encoded + " ", want: true},
		{name: "empty username", username: "", password: encoded, want: false},
		{name: "empty password", username: sncode, password: "", want: false},
		{name: "wrong password", username: sncode, password: "bad-password", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validDeviceCredential(tt.username, tt.password)
			if got != tt.want {
				t.Fatalf("validDeviceCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}
