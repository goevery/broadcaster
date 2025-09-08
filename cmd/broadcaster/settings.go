package main

type Settings struct {
	Port        int      `env:"PORT,default=8000"`
	LogEncoding string   `env:"LOG_ENCODING,default=console"`
	JWTSecret   string   `env:"JWT_SECRET,required=true"`
	APIKeys     []string `env:"API_KEYS,required=true"`
	BasePath    string   `env:"BASE_PATH,default=/broadcaster"`
}
