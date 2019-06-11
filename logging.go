package zapctx

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const unsetValue string = `NOT_SET`

type config struct {
	Environment      string `default:"prod"`
	RedirectStdLog   bool   `default:"true"`
	Level            string `default:"NOT_SET"`
	EnableStackTrace bool   `default:"false"`
	Encoding         string `default:"NOT_SET"`
}

var zapConfig = zap.NewProductionConfig()

func init() {
	ReadLoggingConfig()
}

func ReadLoggingConfig() {
	conf := config{}
	err := envconfig.Process(`LOG`, &conf)
	if err == nil {
		switch strings.ToLower(conf.Environment) {
		case "dev":
			zapConfig = zap.NewDevelopmentConfig()
			zapConfig.Encoding = "console"
		case "qa":
			zapConfig = zap.NewDevelopmentConfig()
		default:
			zapConfig = zap.NewProductionConfig()
		}
		if conf.Encoding != unsetValue {
			zapConfig.Encoding = conf.Encoding
		}

		zapConfig.DisableStacktrace = !conf.EnableStackTrace

		if conf.Level != unsetValue {
			var zapLevel zapcore.Level
			err = zapLevel.UnmarshalText([]byte(conf.Level))
			if err != nil {
				zapLevel = zapcore.InfoLevel
			}

			zapConfig.Level.SetLevel(zapLevel)
		}

		logger, err := zapConfig.Build()
		if err != nil {
			// Default it to production
			logger, _ = zap.NewProduction()
			logger.Error("Failed creating logger with current settings.  Using production defaults.", zap.Any("Config", conf))
		}
		zap.ReplaceGlobals(logger)
		if conf.RedirectStdLog {
			zap.RedirectStdLog(zap.L())
		}
	}
}

func Init(appName string) *zap.Logger {
	ReadLoggingConfig()
	zap.ReplaceGlobals(zap.L().With(zap.String(`app`, appName)))
	return zap.L()
}

func InitTest(t testing.TB, discard bool) *zap.Logger {
	if discard {
		zap.ReplaceGlobals(zap.NewNop())
		return zap.L()
	}
	zapConfig = zap.NewDevelopmentConfig()
	zapConfig.Encoding = "console"

	os.Setenv("LOG_ENVIRONMENT", "dev")
	defer os.Unsetenv("LOG_ENVIRONMENT")

	return Init(t.Name())
}

// WithFields adds all of the given fields to the context logger.
func WithFields(ctx context.Context, fields ...zapcore.Field) context.Context {
	return WithLogger(ctx, Logger(ctx).With(fields...))
}

// Logger returns the zap logger from the given context
func Logger(ctx context.Context) *zap.Logger {
	ctxLogger := ctx.Value(contextKeyLogger)

	l, ok := ctxLogger.(*zap.Logger)
	if !ok {
		return zap.L()
	}
	return l
}

// L returns the global logger for this instance
func L() *zap.Logger {
	return zap.L()
}

// WithLogger will put the specified logger in the returned context.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// WithDefaultLogger will put the specified logger in the returned context.
func WithDefaultLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyLogger, L())
}

// CopyLoggerToContext copies the logger from one context to another.
func CopyLoggerToContext(src, dst context.Context) context.Context {
	ctxLogger := Logger(src)
	return context.WithValue(dst, contextKeyLogger, ctxLogger)
}
