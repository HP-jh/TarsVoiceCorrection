package corrector

import (
	"github.com/HP-jh/TarsVoiceCorrection/internal/opencc"
)

// OpenCCConverter handles Simplified/Traditional Chinese conversion.
type OpenCCConverter struct{}

// NewOpenCCConverter creates a new converter.
func NewOpenCCConverter() (*OpenCCConverter, error) {
	return &OpenCCConverter{}, nil
}

// S2T converts Simplified Chinese to Traditional.
func (occ *OpenCCConverter) S2T(text string) (string, error) {
	return opencc.S2T(text), nil
}

// T2S converts Traditional Chinese to Simplified.
func (occ *OpenCCConverter) T2S(text string) (string, error) {
	return opencc.T2S(text), nil
}

// AutoConvert detects and normalizes mixed Simplified/Traditional text.
// It converts everything to Simplified as the canonical form.
func (occ *OpenCCConverter) AutoConvert(text string) (string, error) {
	return opencc.AutoConvert(text), nil
}
