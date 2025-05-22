package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds common security headers to responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		// A restrictive CSP policy. Adjust as needed for specific frontend requirements if any.
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")
		c.Header("X-XSS-Protection", "1; mode=block") // For older browsers
		// Consider adding other headers like Referrer-Policy, Strict-Transport-Security (if HTTPS is enforced)
		c.Next()
	}
}
