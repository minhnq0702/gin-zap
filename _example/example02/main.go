package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	ginzap "github.com/minhnq0702/gin-zap"
	"github.com/ztrue/tracerr"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	r := gin.New()

	// logger, _ := zap.NewDevelopment()
	logger, _ := zap.NewProduction()
	// encoderCfg := zapcore.EncoderConfig{
	// 	MessageKey:     "msg",
	// 	LevelKey:       "level",
	// 	TimeKey:        "ts",
	// 	NameKey:        "logger",
	// 	CallerKey:      "caller",
	// 	StacktraceKey:  "stacktrace",
	// 	LineEnding:     zapcore.DefaultLineEnding,
	// 	EncodeLevel:    zapcore.CapitalColorLevelEncoder, // color base log level
	// 	EncodeTime:     zapcore.ISO8601TimeEncoder,
	// 	EncodeCaller:   zapcore.ShortCallerEncoder,
	// 	EncodeDuration: zapcore.StringDurationEncoder,
	// }

	// core := zapcore.NewCore(
	// 	zapcore.NewConsoleEncoder(encoderCfg), // init new conosle encoder with color
	// 	zapcore.AddSync(os.Stdout),
	// 	zap.DebugLevel,
	// )

	// logger := zap.New(
	// 	core,
	// 	zap.AddCaller(),
	// 	zap.AddStacktrace(zap.ErrorLevel),
	// )

	r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			fields := []zapcore.Field{}
			// log request ID
			if requestID := c.Writer.Header().Get("X-Request-Id"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}

			// log request body
			var body []byte
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ = io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)
			fields = append(fields, zap.String("body", string(body)))

			return fields
		}),
	}))

	// Example ping request.
	r.GET("/ping", func(c *gin.Context) {
		c.Writer.Header().Add("X-Request-Id", "1234-5678-9012")
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	r.POST("/ping", func(c *gin.Context) {
		c.Writer.Header().Add("X-Request-Id", "9012-5678-1234")
		logger.Debug("request debuging...", zap.String("request_id", "9012-5678-1234"))
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	r.GET("/error", func(c *gin.Context) {
		err := tracerr.New("error message")
		// logger.Error("error message", zap.Error(err))
		// c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
		c.Error(err)
	})

	// Listen and Server in 0.0.0.0:8080
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
