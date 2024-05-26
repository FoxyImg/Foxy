package env

type FoxyEnv struct {
	DatabaseUrl string `env:"DB_URL"`
	RedisUrl    string `env:"REDIS_URL"`

	Port string `env:"PORT"`

	UseCache bool    `env:"USE_CACHE"`
	CacheDir *string `env:"CACHE_DIR"`

	UseRenderCache bool    `env:"USE_RENDER_CACHE"`
	RenderCacheDir *string `env:"RENDER_CACHE_DIR"`

	UseSourceConfigCache bool `env:"USE_SOURCE_CONFIG_CACHE"`

	RequireSignatureValidation bool `env:"REQUIRE_SIG_VALIDATION"`

	DefaultUserName         *string `env:"DEFAULT_USER_NAME"`
	DefaultUserPassword     *string `env:"DEFAULT_USER_PASSWORD"`
	DefaultUserEmail        *string `env:"DEFAULT_USER_EMAIL"`
	DefaultUserKey          *string `env:"DEFAULT_USER_KEY"`
	DefaultUserSecret       *string `env:"DEFAULT_USER_SECRET"`
	DefaultUserSourceConfig *string `env:"DEFAULT_USER_SOURCE_CONFIG_FILE"`
}

var FoxyEnvironment = FoxyEnv{
	Port:                 "8080",
	UseCache:             false,
	UseRenderCache:       false,
	UseSourceConfigCache: true,
}

func Boot() {
	LoadEnvironment(&FoxyEnvironment)
}
