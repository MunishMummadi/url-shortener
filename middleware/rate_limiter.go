package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"url-shortener/logging" // Added for logrus

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus" // Added for logrus fields
)

type visitor struct {
	count    int
	lastSeen time.Time
}

var (
	visitors       = make(map[string]*visitor)
	mu             sync.Mutex
	maxRequests    int
	windowDuration time.Duration
	// rateLimiterLogger = log.New(os.Stdout, "[RateLimiter] ", log.LstdFlags) // Removed
)

func init() {
	var err error
	maxRequestsStr := os.Getenv("MAX_REQUESTS_PER_MINUTE")
	if maxRequests, err = strconv.Atoi(maxRequestsStr); err != nil {
		maxRequests = 40 // default
		logging.Log.WithError(err).WithField("value", maxRequestsStr).Warn("MAX_REQUESTS_PER_MINUTE defaulted")
	}

	windowSecondsStr := os.Getenv("RATE_LIMIT_WINDOW_SECONDS")
	var windowSeconds int
	if windowSeconds, err = strconv.Atoi(windowSecondsStr); err != nil {
		windowSeconds = 60 // default
		logging.Log.WithError(err).WithField("value", windowSecondsStr).Warn("RATE_LIMIT_WINDOW_SECONDS defaulted")
	}
	windowDuration = time.Second * time.Duration(windowSeconds)

	logging.Log.WithFields(logrus.Fields{
		"maxRequests":    maxRequests,
		"windowDuration": windowDuration.String(),
	}).Info("Loaded rate limiter configuration")
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		mu.Lock()
		defer mu.Unlock() // Ensure mutex is always unlocked

		v, exists := visitors[clientIP]
		if !exists {
			visitors[clientIP] = &visitor{
				count:    1,
				lastSeen: time.Now(),
			}
			c.Next()
			return
		}

		// Reset count if window has passed
		if time.Since(v.lastSeen) > windowDuration {
			v.count = 1
			v.lastSeen = time.Now()
		} else if v.count >= maxRequests {
			// mu.Unlock() // Not needed due to defer
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		} else {
			v.count++
		}
		v.lastSeen = time.Now() // Update lastSeen for every request that passes
		c.Next()
	}
}
