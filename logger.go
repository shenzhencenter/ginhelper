package ginhelper

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var skipLoggerPaths *StringSet

func LoggerMiddleware(logger *zap.Logger, skip ...string) gin.HandlerFunc {
	skipLoggerPaths = NewStringSet(skip...)

	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path
		c.Next()

		if !skipLoggerPaths.Contains(path) {

			var fields []zapcore.Field = []zapcore.Field{
				zap.Int("http_status", c.Writer.Status()),
				zap.String("ip", c.ClientIP()),
				zap.String("request.path", path),
				zap.String("request.user_agent", c.Request.UserAgent()),
				zap.String("request.method", c.Request.Method),
				zap.String("request.protocol", c.Request.Proto),
				zap.Int("request.proto_major", c.Request.ProtoMajor),
				zap.Int("request.proto_minor", c.Request.ProtoMinor),
				zap.Any("request.header", c.Request.Header),
				zap.Int64("request.content_length", c.Request.ContentLength),
				zap.Strings("request.transfer_encoding", c.Request.TransferEncoding),
				zap.Bool("request.close", c.Request.Close),
				zap.String("request.host", c.Request.Host),
				zap.Any("request.form", c.Request.Form),
				zap.Any("request.post_form", c.Request.PostForm),
				zap.Any("request.multipart_form", c.Request.MultipartForm),
				zap.Any("request.trailer", c.Request.Trailer),
				zap.String("request.remote_addr", c.Request.RemoteAddr),
				zap.String("request.request_uri", c.Request.RequestURI),
				zap.String("request.fullpath", c.FullPath()),
			}

			end := time.Now().UTC()
			latency := end.Sub(start)
			fields = append(fields, zap.Duration("latency", latency), zap.Time("created_at", start), zap.Time("finished_at", end))

			if c.Params != nil {
				fields = append(fields, zap.Any("params", c.Params))
			}

			if len(c.Errors) > 0 {
				for _, e := range c.Errors.Errors() {
					logger.Error(e, fields...)
				}
			} else {
				logger.Info(path, fields...)
			}
		}
	}
}

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return handleRecover(logger, func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": "internal_server_error", "msg": "Internal Server Error"})
	})
}

func handleRecover(logger *zap.Logger, recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, true)
				if brokenPipe {
					logger.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				logger.Error("recover_from_panic",
					zap.Time("time", time.Now()),
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
					zap.String("stack", string(debug.Stack())),
				)
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
