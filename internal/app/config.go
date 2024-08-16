package app

type ServerConfig struct {
	Addr string `envconfig:"SERVER_ADDR" default:":9000"`
}

type WeatherConfig struct {
	APIKey string
}

type GitHubConfig struct {
	Token string `envconfig:"GITHUB_TOKEN"`
}

type Config struct {
	Server  ServerConfig
	Weather WeatherConfig
	GitHub  GitHubConfig
}
