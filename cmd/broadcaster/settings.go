package main

type Settings struct {
	Port       int    `env:"PORT,default=8000"`
	MongoDbUri string `env:"MONGODB_URI,required=true"`
	JWTSecret  string `env:"JWT_SECRET,required=true"`
	BasePath   string `env:"BASE_PATH,default=/broadcaster"`
}
