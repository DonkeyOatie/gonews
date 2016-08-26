package main

import (
	"log"

	"gopkg.in/redis.v4"
)

const redisXMLListName = "NEWS_XML"
const redisKeySetName = "NEWS_KEYS"

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

// storeNewsPost stores the string representation of the XML news item in the
// list NEWS_XML in redis.  We use RPUSH to add to the end of the list
func storeNewsPost(post string) {
	rd.RPush(redisXMLListName, post)
}

// storeNewsKey adds the unixEpoch from the file name to the key SET in redis
func storeNewsKey(key string) {
	rd.SAdd(redisKeySetName, key)
}

// isKeyPresent checks to see if the unixEpoch from the archive file name is
// available in the SET redisKeySetName in redis.  Returns true of the key is
// there, false if not
func isKeyPresent(key string) bool {
	scan := rd.SScan(redisKeySetName, 0, key, 0)
	keys, _, _ := scan.ScanCmd.Result()
	return len(keys) > 0
}
