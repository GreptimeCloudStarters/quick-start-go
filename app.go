package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
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
var port = flag.String("port", "", "The port of the HTTP endpoint of GreptimeDB.")
var db = flag.String("db", "public", "The name of the database of GreptimeDB.")
var username = flag.String("username", "", "The username of the database.")
var password = flag.String("password", "", "The password of the database.")
var endpoint = flag.String("endpoint", "", "The HTTP endpoint of OTLP/HTTP exporter.")

func main() {
	flag.Parse()

	opts, err := generateOtlpHttpOptionsFromEndpoint(*endpoint)

	if err != nil {
		panic(err)
	}

	if opts == nil {
		opts, err = generateOtlpHttpOptionsFromHost(*dbHost, *port)
		if err != nil {
			panic(err)
		}
	}

	opts = append(opts, otlpmetrichttp.WithTimeout(time.Second*5))
	headers := generateOtlpHttpHeaders(*db, *username, *password)
	opts = append(opts, otlpmetrichttp.WithHeaders(headers))

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

func generateOtlpHttpOptionsFromEndpoint(endpoint string) ([]otlpmetrichttp.Option, error) {
	if endpoint == "" {
		return nil, nil
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	opts := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(u.Host)}
	opts = append(opts, otlpmetrichttp.WithURLPath(u.Path))
	if u.Scheme == "http" {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	return opts, nil
}

func generateOtlpHttpOptionsFromHost(host string, port string) ([]otlpmetrichttp.Option, error) {
	if host == "" {
		return nil, errors.New("endpoint url or host is required")
	}
	opts := []otlpmetrichttp.Option{otlpmetrichttp.WithURLPath("/v1/otlp/v1/metrics")}
	endpoint := host
	if port != "" {
		endpoint = fmt.Sprintf("%s:%s", host, port)
	}
	opts = append(opts, otlpmetrichttp.WithEndpoint(endpoint))
	return opts, nil
}

func generateOtlpHttpHeaders(db string, username string, password string) map[string]string {
	headers := map[string]string{"x-greptime-db-name": db}
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		headers["Authorization"] = "Basic " + auth
	}
	return headers
}
