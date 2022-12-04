package telemetry_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/su-starter-kit/telemetry/telemetry_metric"
	"github.com/su-starter-kit/telemetry/telemetry_trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/trace"
	otel_trace "go.opentelemetry.io/otel/trace"
)

func Example_Telemetry_TraceProvider() {
	// Creates a console trace provider
	todoContext := context.TODO()

	tp, err := telemetry_trace.NewTracerProvider(
		telemetry_trace.WithResource("test_service", "v0.0.0", "dev"),
		telemetry_trace.WithConsoleExporter(os.Stdout),
		telemetry_trace.WithSampler(trace.AlwaysSample()),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tp.Shutdown(todoContext); err != nil {
			log.Fatal(err)
		}
	}()

	otel.SetTracerProvider(tp)
	tcr := otel.Tracer("my_service")
	// This operation shows an example on how to propagate spans accross
	// operations
	parentOperation(todoContext, tcr)
}

func Example_Telemetry_MeterProvider() {
	todoContext := context.TODO()
	// Creates a console meter provider
	mp, err := telemetry_metric.NewMeterProvider(
		telemetry_metric.WithConsoleExporter(os.Stdout),
		telemetry_metric.WithResource("test_service", "v0.0.0", "dev"),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := mp.Shutdown(todoContext); err != nil {
			log.Fatal(err)
		}
	}()

	global.SetMeterProvider(mp)
	mtr := global.Meter("my_service")

	counter, _ := mtr.SyncFloat64().Counter("my_counter")
	gauge, _ := mtr.SyncFloat64().UpDownCounter("my_gauge")
	histogram, _ := mtr.SyncFloat64().Histogram("my_histogram")

	counter.Add(todoContext, 1)
	gauge.Add(todoContext, 50)
	histogram.Record(todoContext, 2)
}

// ----------------------------------------------------------------
// Functions behind are used to demonstrate the span propagation
// and span metadata attachement options.
// ----------------------------------------------------------------

func parentOperation(ctx context.Context, tcr otel_trace.Tracer) {
	// Starts a time span
	parentCtx, firstSpan := tcr.Start(
		ctx,
		"parent_operation",
		// This attributes will be appended as part of the span metadata
		// in form of tags, which can be used to filter in queries and dashboards
		// in your trace system console, or through Grafana.
		otel_trace.WithAttributes(
			attribute.String("vendor_name", "some vendor name"),
			attribute.Int("quantity_required", 2),
		),
	)
	defer firstSpan.End()

	fmt.Println("First operation!")
	// Calls a child operation passing the context in order to propagate the span to child operation
	childOperationNested(parentCtx, tcr)
}

func childOperationNested(parentCtx context.Context, tcr otel_trace.Tracer) {
	// Uses the parentCtx context in order to propagate the parent span
	firstChildContext, nestedSpan := tcr.Start(parentCtx, "nested_operation")
	defer nestedSpan.End()
	fmt.Println("First child operation!")
	// This operation will fail,
	// in order to propagate the failure through the span one can use
	// the opetions used below:
	if err := childFaultyOperationNested(firstChildContext, tcr); err != nil {
		// Following operations with spans are recommended
		// since it leads to a better debug
		nestedSpan.SetStatus(
			codes.Error,
			err.Error(),
		)
		nestedSpan.RecordError(
			err,
			otel_trace.WithStackTrace(true),
		)

		// One can get span data (TraceID and SpanID) to be propagated
		// within trasanction message, objects, and logs
		spanData := telemetry_trace.GetContextData(nestedSpan)
		fmt.Printf("%v\n", spanData)
	}
}

func childFaultyOperationNested(parentCtx context.Context, tcr otel_trace.Tracer) error {
	_, nestedSpan := tcr.Start(parentCtx, "second_nested_operatioin")
	defer nestedSpan.End()
	fmt.Println("Faulty child span was called!")

	// One can add events to the span also
	nestedSpan.AddEvent("faulty_span_was_called")

	return fmt.Errorf("This operation are presenting some defect.")
}
