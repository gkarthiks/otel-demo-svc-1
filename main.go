package main

import (
	"context"
	"fmt"
	discovery "github.com/gkarthiks/k8s-discovery"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
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
	http.Handle("/pods", NewHandler(http.HandlerFunc(listPods), "list pods"))

	http.ListenAndServe(":8090", nil)
}

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong\n")
}

func listPods(w http.ResponseWriter, req *http.Request) {
	k8s, _ := discovery.NewK8s()
	var pods []string

	podList, _ := k8s.Clientset.CoreV1().Pods("").List(context.Background(),
		v1.ListOptions{})

	for _, pod := range podList.Items {
		pods = append(pods, pod.Name)
	}

	fmt.Printf("Pods in the cluster: %v", pods)
}