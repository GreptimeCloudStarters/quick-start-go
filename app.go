package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	appHost "go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var dbHost = flag.String("host", "localhost", "The host address of the GreptimeCloud service")
var db = flag.String("db", "public", "The name of the database of the GreptimeCloud service")
var username = flag.String("username", "greptimeUser", "The username of the database")
var password = flag.String("password", "greptimePassword", "The password of the database")

func main() {
	flag.Parse()
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *username, *password)))

	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpoint(*dbHost),
		otlpmetrichttp.WithURLPath("/v1/otlp/v1/metrics"),
		otlpmetrichttp.WithHeaders(map[string]string{
			"x-greptime-db-name": *db,
			"Authorization":      "Basic " + auth,
		}),
		otlpmetrichttp.WithTimeout(time.Second*5),
	)

	if err != nil {
		panic(err)
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second*2))

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("quick-start-demo-go"),
		semconv.ServiceVersion("v0.1.0"),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
	)

	defer func() {
		err := meterProvider.Shutdown(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Print("Sending metrics...")
	err = appHost.Start(appHost.WithMeterProvider(meterProvider))
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
