package env

type FoxyEnv struct {
	Isolated bool

	ServerType  string  `env:"SERVER_TYPE"`
	DatabaseUrl *string `env:"DB_URL"`
	RedisUrl    *string `env:"REDIS_URL"`

	SourceConfigFile *string `env:"SOURCE_CONFIG"`
	PresetsFile      *string `env:"PRESETS"`

	Port string `env:"PORT"`

	UseCache bool    `env:"USE_CACHE"`
	CacheDir *string `env:"CACHE_DIR"`

	MaxSourceSize int `env:"MAX_SOURCE_SIZE"`

	UseRenderCache bool    `env:"USE_RENDER_CACHE"`
	RenderCacheDir *string `env:"RENDER_CACHE_DIR"`

	UseSourceConfigCache bool `env:"USE_SOURCE_CONFIG_CACHE"`

	RequireSignatureValidation bool `env:"REQUIRE_SIG_VALIDATION"`

	AllowPresetManagement bool `env:"ALLOW_PRESET_MANAGEMENT"`
}

var FoxyEnvironment = FoxyEnv{
	ServerType:            "primary",
	Port:                  "8080",
	UseCache:              false,
	UseRenderCache:        false,
	UseSourceConfigCache:  true,
	AllowPresetManagement: false,
	MaxSourceSize:         3840,
}

func Boot() {
	LoadEnvironment(&FoxyEnvironment)
	FoxyEnvironment.Isolated = FoxyEnvironment.SourceConfigFile != nil
	if FoxyEnvironment.Isolated {
		FoxyEnvironment.DatabaseUrl = nil
	}
}
