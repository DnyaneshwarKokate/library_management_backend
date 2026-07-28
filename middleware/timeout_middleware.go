package middleware

import (
	"context"
	"errors"
	"time"

	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{}, 1)
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicChan <- r
				}
			}()
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			return
		case p := <-panicChan:
			logrus.Errorf("TimeoutMiddleware panic recovered: %v", p)
			panic(p)
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logrus.Warnf("Request timed out after %v: %s %s", timeout, c.Request.Method, c.Request.URL.Path)
				utils.GatewayTimeoutAbortWithJSON(c, "Request timeout: processing took too long")
			}
		}
	}
}
