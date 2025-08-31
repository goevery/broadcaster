package main

type Settings struct {
	Port       int    `env:"PORT,default=8000"`
	MongoDbUri string `env:"MONGODB_URI,required=true"`
}
