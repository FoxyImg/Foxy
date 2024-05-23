package config

type S3Config struct {
	Key    *string `json:"key"`
	Secret *string `json:"secret"`
	Bucket *string `json:"bucket"`
	Region *string `json:"region"`
}

type SourceConfig struct {
	Type string `json:"type"`
	S3Config
}

type VisionConfig struct {
	Enabled              bool   `json:"enabled"`
	Type                 string `json:"type"`
	UseSourceCredentials bool   `json:"useSourceCredentials"`
	S3Config
}

type Config struct {
	Secret *string       `json:"secret"`
	Source *SourceConfig `json:"source"`
	Vision *VisionConfig `json:"vision"`
}
