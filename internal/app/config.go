package app

type WeatherConfig struct {
	APIKey string `yaml:"api_key"`
}

type AlertsInUAConfig struct {
	APIKey string `yaml:"api_key"`
}

type GitHubConfig struct {
	Token string `yaml:"token"`
}

type Config struct {
	Weather    WeatherConfig    `yaml:"weather"`
	AlertsInUA AlertsInUAConfig `yaml:"alerts_in_ua"`
	GitHub     GitHubConfig     `yaml:"github"`
}
