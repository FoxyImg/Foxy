package env

import (
	"foxy/internal/utils"
	"log"
	"time"
)

type FoxyEnv struct {
	Isolated bool

	ServerType  string  `env:"SERVER_TYPE"`
	DatabaseUrl *string `env:"DB_URL"`
	RedisUrl    *string `env:"REDIS_URL"`

	SourceConfigFile *string `env:"SOURCE_CONFIG"`
	PresetsFile      *string `env:"PRESETS"`

	Port           string `env:"PORT"`
	IdleTimeout    *int   `env:"IDLE_TIMEOUT"`
	ReadTimeout    *int   `env:"READ_TIMEOUT"`
	WriteTimeout   *int   `env:"WRITE_TIMEOUT"`
	MaxConnections *int   `env:"MAX_CONNECTIONS"`

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

	UseRateLimiter  *bool `env:"USE_RATE_LIMITER"`
	TokensPerMinute *int  `env:"TOKENS_PER_MINUTE"`

	Concurrency *int `env:"CONCURRENCY"`
	MaxBurst    *int `env:"MAX_BURST"`

	CacheTTL *time.Duration `env:"CACHE_TTL"`
}

var FoxyEnvironment = FoxyEnv{
	ServerType:            "primary",
	Port:                  "8080",
	IdleTimeout:           utils.Ptr(60),
	ReadTimeout:           utils.Ptr(1),
	WriteTimeout:          utils.Ptr(10),
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
	UseRateLimiter:        utils.Ptr(false),
	TokensPerMinute:       utils.Ptr(30),
	MaxConnections:        utils.Ptr(0),
	Concurrency:           utils.Ptr(0),
	MaxBurst:              utils.Ptr(100),
	CacheTTL:              utils.Ptr(0 * time.Second),
}

func Boot() {
	LoadEnvironment(&FoxyEnvironment)
	FoxyEnvironment.Isolated = FoxyEnvironment.SourceConfigFile != nil
	if FoxyEnvironment.Isolated {
		FoxyEnvironment.DatabaseUrl = nil
	}

	log.Printf("ServerType: %s", FoxyEnvironment.ServerType)
	log.Printf("SourceConfigFile: %s", *FoxyEnvironment.SourceConfigFile)
	log.Printf("PresetsFile: %s", *FoxyEnvironment.PresetsFile)
	log.Printf("Port: %s", FoxyEnvironment.Port)
	log.Printf("IdleTimeout: %d", *FoxyEnvironment.IdleTimeout)
	log.Printf("ReadTimeout: %d", *FoxyEnvironment.ReadTimeout)
	log.Printf("WriteTimeout: %d", *FoxyEnvironment.WriteTimeout)
	log.Printf("MaxConnections: %d", *FoxyEnvironment.MaxConnections)
	log.Printf("UseCache: %t", FoxyEnvironment.UseCache)
	log.Printf("UseVisionCache: %t", FoxyEnvironment.UseVisionCache)
	log.Printf("UseRenderCache: %t", FoxyEnvironment.UseRenderCache)
	log.Printf("UseSourceConfigCache: %t", FoxyEnvironment.UseSourceConfigCache)
	log.Printf("AllowPresetManagement: %t", FoxyEnvironment.AllowPresetManagement)
	log.Printf("MaxSourceSize: %d", FoxyEnvironment.MaxSourceSize)
	log.Printf("AlwaysPrerender: %t", FoxyEnvironment.AlwaysPrerender)
	log.Printf("DebugImages: %t", *FoxyEnvironment.DebugImages)
	log.Printf("UseML: %t", *FoxyEnvironment.UseML)
	log.Printf("OnnxUseCoreML: %t", *FoxyEnvironment.OnnxUseCoreML)
	log.Printf("UseRateLimiter: %t", *FoxyEnvironment.UseRateLimiter)
	log.Printf("TokensPerMinute: %d", *FoxyEnvironment.TokensPerMinute)
	log.Printf("Concurrency: %d", *FoxyEnvironment.Concurrency)
	log.Printf("MaxBurst: %d", *FoxyEnvironment.MaxBurst)
	log.Printf("CacheTTL: %d", FoxyEnvironment.CacheTTL.String())
}
