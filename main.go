package main

// From https://gobyexample.com/http-servers

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	storage "cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

func push(wr http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println(err)
	}
	bkt := client.Bucket(os.Getenv("BUCKET_NAME"))
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println(err)
	}
	sha256 := sha256.Sum256(body)
	obj := bkt.Object(fmt.Sprintf("%x", sha256))
	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, string(body[:])); err != nil {
		fmt.Println(err)
	}
	if err := w.Close(); err != nil {
		fmt.Println(err)
	}
}

func healthz(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok\n")
	fmt.Println("healthz called")
}

func fail(w http.ResponseWriter, req *http.Request) {
	fmt.Println("fail called")
	_, exists := os.LookupEnv("SENTRY_DSN")
	if exists {
		sentry.CaptureException(errors.New("/fail called")) // check Sentry
	}
	panic("fail") // check Google Cloud Error Reporting
}

func main() {
	value, exists := os.LookupEnv("SENTRY_DSN")
	if exists {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              value,
			Debug:            true,
			TracesSampleRate: 1.0,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		defer sentry.Flush(2 * time.Second)
		fmt.Println("Sentry initialized")
		sentry.CaptureMessage("check")
	}
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// carefull
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	http.HandleFunc("/push", push)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/fail", fail)

	fmt.Println("Listening on port", os.Getenv("PORT"))
	err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
	if err != nil {
		panic(err)
	}
}
