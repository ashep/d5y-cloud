package fwupdate

import "testing"

func TestIsBootloaderAsset(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"cronus-0.0.5-esp32s3_ws2812v32x16.bin", false},
		{"cronus-0.0.6-alpha1-esp32s3_max7219v32x16.bin", false},
		{"cronus-bootloader-0.0.5-esp32s3_ws2812v32x16.bin", true},
		{"cronus-bootloader-0.0.6-alpha1-esp32s3_max7219v32x16.bin", true},
		{"other-bootloader-0.0.5-esp32s3_ws2812v32x16.bin", false},
	}

	for _, tt := range tests {
		if got := isBootloaderAsset(tt.name, "cronus"); got != tt.want {
			t.Errorf("isBootloaderAsset(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
