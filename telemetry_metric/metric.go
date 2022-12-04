package telemetry_metric

import (
	"context"
	"encoding/json"
	"io"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type meterProviderOption func(*[]metric.Option) error

func NewMeterProvider(opts ...meterProviderOption) (*metric.MeterProvider, error) {
	var meterProviderOptions []metric.Option = make([]metric.Option, 0)
	for _, op := range opts {
		if err := op(&meterProviderOptions); err != nil {
			return nil, err
		}
	}

	return metric.NewMeterProvider(meterProviderOptions...), nil
}

func WithOtlpHttpCollectorExporter(ctx context.Context, endpoint string) meterProviderOption {
	return func(otelOptions *[]metric.Option) error {
		exporter, err := otlpmetrichttp.New(
			ctx,
			otlpmetrichttp.WithEndpoint(endpoint),
			otlpmetrichttp.WithInsecure(),
		)
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, metric.WithReader(metric.NewPeriodicReader(exporter)))
		return nil
	}
}

func WithConsoleExporter(w io.Writer) meterProviderOption {
	return func(otelOptions *[]metric.Option) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "	")
		exporter, err := stdoutmetric.New(
			stdoutmetric.WithEncoder(enc),
		)
		if err != nil {
			return err
		}

		*otelOptions = append(*otelOptions, metric.WithReader(metric.NewPeriodicReader(exporter)))
		return nil
	}
}

func WithResource(serviceName, serviceVersion, environment string) meterProviderOption {
	return func(otelOptions *[]metric.Option) error {
		r := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", environment),
		)

		*otelOptions = append(*otelOptions, metric.WithResource(r))
		return nil
	}
}

// WithOpentelemetryMetricOption
// Bypasses options for Metric Provider Builder and sets
// opentelemetry options directly
func WithOpentelemetryMetricOption(opts []metric.Option) meterProviderOption {
	return func(otelOptions *[]metric.Option) error {
		for _, opt := range opts {
			*otelOptions = append(*otelOptions, opt)
		}
		return nil
	}
}
