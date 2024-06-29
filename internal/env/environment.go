package env

import (
	"foxy/internal/utils"
	"time"
)

type FoxyEnv struct {
	Isolated bool

	ServerType  string  `env:"SERVER_TYPE"`
	DatabaseUrl *string `env:"DB_URL"`
	RedisUrl    *string `env:"REDIS_URL"`

	SourceConfigFile *string `env:"SOURCE_CONFIG"`
	PresetsFile      *string `env:"PRESETS"`

	Port         string        `env:"PORT"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT"`

	UseCache       bool    `env:"USE_CACHE"`
	UseVisionCache bool    `env:"USE_VISION_CACHE"`
	CacheDir       *string `env:"CACHE_DIR"`

	MaxSourceSize int `env:"MAX_SOURCE_SIZE"`

	UseRenderCache bool    `env:"USE_RENDER_CACHE"`
	RenderCacheDir *string `env:"RENDER_CACHE_DIR"`

	UseSourceConfigCache bool `env:"USE_SOURCE_CONFIG_CACHE"`

	RequireSignatureValidation bool `env:"REQUIRE_SIG_VALIDATION"`
	RequirePresetSignature     bool `env:"REQUIRE_PRESET_SIG_VALIDATION"`

	AllowPresetManagement bool    `env:"ALLOW_PRESET_MANAGEMENT"`
	APIKey                *string `env:"API_KEY"`

	AlwaysPrerender bool `env:"ALWAYS_PRERENDER"`

	DebugImages *bool `env:"DEBUG_IMAGES"`

	UseML               *bool   `env:"USE_ML"`
	OnnxLib             *string `env:"ONNX_LIB"`
	OnnxModelHumans     *string `env:"ONNX_MODEL_HUMANS"`
	OnnxModelForeground *string `env:"ONNX_MODEL_FOREGROUND"`
	OnnxUseCoreML       *bool   `env:"ONNX_USE_CORE_ML"`
}

var FoxyEnvironment = FoxyEnv{
	ServerType:            "primary",
	Port:                  "8080",
	IdleTimeout:           60 * time.Second,
	ReadTimeout:           1 * time.Second,
	WriteTimeout:          10 * time.Second,
	UseCache:              false,
	UseVisionCache:        false,
	UseRenderCache:        false,
	UseSourceConfigCache:  true,
	AllowPresetManagement: false,
	MaxSourceSize:         3840,
	AlwaysPrerender:       true,
	DebugImages:           utils.Ptr(false),
	UseML:                 utils.Ptr(false),
	OnnxUseCoreML:         utils.Ptr(false),
}

func Boot() {
	LoadEnvironment(&FoxyEnvironment)
	FoxyEnvironment.Isolated = FoxyEnvironment.SourceConfigFile != nil
	if FoxyEnvironment.Isolated {
		FoxyEnvironment.DatabaseUrl = nil
	}
}
