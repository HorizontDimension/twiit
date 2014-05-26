package twiit

import (
	"github.com/robfig/go-cache"
	"time"
)

//Global TwiitCache
// Create a cache with a default expiration time of 5 minutes, and which
// purges expired items every 30 seconds
var Cache = cache.New(5*time.Minute, 30*time.Second)

var CacheAuth = cache.New(5*time.Hour, 30*time.Second)
