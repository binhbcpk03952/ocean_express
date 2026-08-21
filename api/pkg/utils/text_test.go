package utils_test

import (
	"ocean-express-api/pkg/utils"
	"testing"
)

func TestFixMojibake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Direct Mojibake Hà Huy Tập",
			input:    "300/6 HÃ  Huy Táº\u00adp",
			expected: "300/6 Hà Huy Tập",
		},
		{
			name:     "Mixed Mojibake address with clean Vietnamese suffix",
			input:    "300/6 HÃ  Huy Táº\u00adp, Phường Tân An, Tỉnh Đắk Lắk",
			expected: "300/6 Hà Huy Tập, Phường Tân An, Tỉnh Đắk Lắk",
		},
		{
			name:     "Clean Vietnamese address",
			input:    "300 HÀ HUY TẬP, Xã Bình Thuận, Tỉnh Sơn La",
			expected: "300 HÀ HUY TẬP, Xã Bình Thuận, Tỉnh Sơn La",
		},
		{
			name:     "Clean name with accents",
			input:    "Bùi Chí Bình",
			expected: "Bùi Chí Bình",
		},
		{
			name:     "English text",
			input:    "Ocean Express Shop",
			expected: "Ocean Express Shop",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.FixMojibake(tc.input)
			if got != tc.expected {
				t.Errorf("FixMojibake(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
