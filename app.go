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

const Version = "1.0.2"

var dbHost = flag.String("host", "localhost", "The host address of GreptimeDB.")
var db = flag.String("db", "public", "The name of the database of GreptimeDB.")
var username = flag.String("username", "", "The username of the database.")
var password = flag.String("password", "", "The password of the database.")

func main() {
	flag.Parse()

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithURLPath("/v1/otlp/v1/metrics"),
		otlpmetrichttp.WithTimeout(time.Second * 5)}

	if *dbHost == "localhost" || *dbHost == "127.0.0.1" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
		opts = append(opts, otlpmetrichttp.WithEndpoint(fmt.Sprintf("%s:4000", *dbHost)))
	} else {
		opts = append(opts, otlpmetrichttp.WithEndpoint(*dbHost))
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *username, *password)))
	opts = append(opts, otlpmetrichttp.WithHeaders(map[string]string{
		"x-greptime-db-name": *db,
		"Authorization":      "Basic " + auth}))

	exporter, err := otlpmetrichttp.New(
		context.Background(),
		opts...,
	)

	if err != nil {
		panic(err)
	}

	reader := metric.NewPeriodicReader(exporter, metric.WithInterval(time.Second*5))

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("quick-start-demo-go"),
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
