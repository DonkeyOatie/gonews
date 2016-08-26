package main

import (
	"log"

	"gopkg.in/redis.v4"
)

const redisXMLListName = "NEWS_XML"

var rd = initRedis()

// initRedis connects to our redis instance and return a connection handler
func initRedis() *redis.Client {
	redisHost := "localhost"
	redisPort := "6379"

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "",
		DB:       0,
	})

	if _, err := client.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	return client
}
