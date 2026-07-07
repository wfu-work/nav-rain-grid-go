package configs

const DefaultVersionPath = "./data/version"

type RainConfig struct {
	VersionPath string `mapstructure:"version-path" json:"versionPath" yaml:"version-path"`
}
