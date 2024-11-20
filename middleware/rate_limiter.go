package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

const (
	maxRequestsPerMinute = 40
)

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		mu.Lock()
		v, exists := visitors[clientIP]
		if !exists {
			visitors[clientIP] = &visitor{
				count:    1,
				lastSeen: time.Now(),
			}
			mu.Unlock()
			c.Next()
			return
		}

		// Reset count if minute has passed
		if time.Since(v.lastSeen) > time.Minute {
			v.count = 1
			v.lastSeen = time.Now()
		} else if v.count >= maxRequestsPerMinute {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		} else {
			v.count++
		}
		v.lastSeen = time.Now()
		mu.Unlock()
		c.Next()
	}
}
