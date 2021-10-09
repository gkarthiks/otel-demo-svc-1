package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"log"
	"net/http"
	"strings"
)

func initTracer() *sdktrace.TracerProvider {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("Service-1"))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func NewHandler(h http.Handler, operation string) http.Handler {
	httpOptions := []otelhttp.Option{
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	}
	return otelhttp.NewHandler(h, operation, httpOptions...)
}

func main() {
	tp := initTracer()

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	http.Handle("/ping", NewHandler(http.HandlerFunc(ping), "ping"))
	http.Handle("/po", NewHandler(http.HandlerFunc(listPods), "list pods"))

	http.ListenAndServe(":8090", nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func listPods(w http.ResponseWriter, req *http.Request) {
	//url := "http://service2.cwitc.svc.cluster.local:80/pods"
	url := "http://localhost:8091/pods"

	ctx := context.Background()
	resp, _ := otelhttp.Get(ctx, url)

	var data []string
	dec := gob.NewDecoder(resp.Body)
	dec.Decode(&data)

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Printf("Pods in cluster: %v", strings.Join(data, "\n"))
	fmt.Fprintf(w, strings.Join(data, ","))
}