package app

type ServerConfig struct {
	Addr string `default:":9000"`
}

type WeatherConfig struct {
	APIKey string
}

type GitHubConfig struct {
	Token string
}

type Config struct {
	Server  ServerConfig
	Weather WeatherConfig
	GitHub  GitHubConfig
}
