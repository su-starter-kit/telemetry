package telemetry_trace

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otel_trace "go.opentelemetry.io/otel/trace"
)

type tracerProviderOption func(*[]trace.TracerProviderOption) error

func NewTracerProvider(opts ...tracerProviderOption) (*trace.TracerProvider, error) {
	var openTelemetryOptions []trace.TracerProviderOption = make([]trace.TracerProviderOption, 0)
	for _, op := range opts {
		if err := op(&openTelemetryOptions); err != nil {
			return nil, err
		}
	}

	return trace.NewTracerProvider(openTelemetryOptions...), nil
}

func WithConsoleExporter(w io.Writer) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithWriter(w),
			// User human readable output
			stdouttrace.WithPrettyPrint(),
			// Do not print timestamps
			stdouttrace.WithoutTimestamps(),
		)
		if err != nil {
			return err
		}
		*otelOptions = append(*otelOptions, trace.WithBatcher(stdoutExporter))

		return nil
	}
}

func WithZipkinExporter(collectorUrl string) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		exporter, err := zipkin.New(collectorUrl)
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, trace.WithBatcher(exporter))
		return nil
	}
}

func WithJaegerExporter(jaegerUrl string) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		jExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, trace.WithBatcher(jExporter))
		return nil
	}
}

func WithOtlpHttpCollectorExporter(ctx context.Context, otlpCollectorExporterUrl string) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {

		client := otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(otlpCollectorExporterUrl),
			otlptracehttp.WithInsecure(),
		)
		exporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, trace.WithBatcher(exporter))
		return nil
	}
}

func WithOtlpGrpcCollectorExporter(ctx context.Context, otlpCollectorExporterUrl string, reconnectionSeconds time.Duration) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {

		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(otlpCollectorExporterUrl),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithReconnectionPeriod(reconnectionSeconds*time.Second),
		)
		exporter, err := otlptrace.New(ctx, client)
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, trace.WithBatcher(exporter))
		return nil
	}
}

func WithResource(serviceName, serviceVersion, environment string) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		r := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", environment),
		)

		*otelOptions = append(*otelOptions, trace.WithResource(r))
		return nil
	}
}

func WithSampler(sampler trace.Sampler) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		*otelOptions = append(*otelOptions, trace.WithSampler(sampler))

		return nil
	}
}

// WithOpenTelemetryTracerProviderOption
// Bypasses options for Trace Provider Builder and sets
// opentelemetry options directly
func WithOpenTelemetryTracerProviderOption(opt trace.TracerProviderOption) tracerProviderOption {
	return func(otelOptions *[]trace.TracerProviderOption) error {
		*otelOptions = append(*otelOptions, opt)

		return nil
	}
}

type SpanContextData struct {
	SpanId  string
	TraceId string
}

func GetContextData(span otel_trace.Span) *SpanContextData {
	return &SpanContextData{
		SpanId:  span.SpanContext().SpanID().String(),
		TraceId: span.SpanContext().TraceID().String(),
	}
}
