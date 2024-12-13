package app

type WeatherConfig struct {
	APIKey string
}

type GitHubConfig struct {
	Token string
}

type Config struct {
	Weather WeatherConfig
	GitHub  GitHubConfig
}
