package config

type Config struct {
	Hosts map[string]Host `yaml:"hosts"`
}

type Host struct {
	Cmd   []string `yaml:"cmd"`   // e.g. ["docker", "exec", "-i", "host1", "norouter"]
	VIP   string   `yaml:"vip"`   // e.g. "127.0.42.101"
	Ports []string `yaml:"ports"` // e.g. ["8080:127.0.0.1:80"], or ["8080:127.0.0.1:80/tcp"]
}
