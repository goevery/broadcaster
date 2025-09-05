package main

type Settings struct {
	Port      int    `env:"PORT,default=8000"`
	JWTSecret string `env:"JWT_SECRET,required=true"`
	BasePath  string `env:"BASE_PATH,default=/broadcaster"`
}
