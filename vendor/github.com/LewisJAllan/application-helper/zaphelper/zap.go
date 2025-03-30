package zaphelper

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapLogger = zap.New(zapcore.NewCore(
	zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:       "msg",
		LevelKey:         "level",
		TimeKey:          "@timestamp",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      "function",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		EncodeTime:       zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration:   zapcore.NanosDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: "",
	}),
	zapcore.AddSync(os.Stdout),
	zap.NewAtomicLevelAt(zapcore.InfoLevel),
),
	zap.AddCaller(),
	zap.AddCallerSkip(1),
)

// FromContext will return the logger associated with the context if present, otherwise the ZapLogger
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(struct{}{}).(*zap.Logger); ok {
		return l
	}
	return ZapLogger
}

// With will add l to the context
func With(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, struct{}{}, logger)
}

// Error will log at the error level using the associated context
func Error(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}

// Info will log at the info level using the associated context
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// Warn will log at the warn level using the associated context
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// Debug will log at the debug level using the associated context
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log entries.  Applications should take care to call
// Sync before exiting.
func Sync(ctx context.Context) error {
	return FromContext(ctx).Sync()
}
