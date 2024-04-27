package config

type Redis struct {
	Db   int
	Host string
	Port string
}

type App struct {
	Host string
	Port string
}

type RateLimiter struct {
	ByIp LimitValues
}

type LimitValues struct {
	MaxReq        int
	TimeWindow    int64
	BlockDuration int64
}

type Config struct {
	Redis       Redis
	App         App
	RateLimiter RateLimiter
}
