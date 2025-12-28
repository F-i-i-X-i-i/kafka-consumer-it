package tracing

import (
	"context"
	"io"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// Init initializes the OpenTelemetry tracer
func Init(serviceName string, w io.Writer) (func(context.Context) error, error) {
	// Create exporter (stdout for development, can be replaced with Jaeger/OTLP)
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource with service info
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer
	tracer = tp.Tracer(serviceName)

	return tp.Shutdown, nil
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracer.Start(ctx, name, opts...)
}

// AddEvent adds an event to the current span
func AddEvent(ctx context.Context, name string, attrs ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, attrs...)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}

// SetAttributes sets attributes on the current span
func SetAttributes(ctx context.Context, attrs ...trace.SpanOption) {
	// Note: This is a simplified version. In real usage, use trace.Span.SetAttributes
	span := trace.SpanFromContext(ctx)
	_ = span // Use the span for actual attribute setting
}

// TraceID returns the trace ID from context
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().TraceID().String()
}

// SpanID returns the span ID from context
func SpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	return span.SpanContext().SpanID().String()
}
