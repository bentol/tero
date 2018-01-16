package config

type Config struct {
	ProxyHost        string `toml:"proxy_host"`
	EnableEmailToken bool   `toml:"enable_email_token"`
	SMTP             SMTPConfig
}

type SMTPConfig struct {
	Host       string
	Username   string
	Password   string
	Sender     string
	SenderName string `toml:"sender_name"`
	Port       int
}

var conf Config

func Set(newConfig Config) {
	conf = newConfig
}

func Get() Config {
	return conf
}
