package main

// From https://gobyexample.com/http-servers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	storage "cloud.google.com/go/storage"
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
	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		fmt.Println(err)
	}
}

func healthz(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok\n")
	fmt.Println("healthz called")
}

func main() {
	http.HandleFunc("/push", push)
	http.HandleFunc("/healthz", healthz)

	fmt.Println("Listening on port", os.Getenv("PORT"))
	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
	if err != nil {
		panic(err)
	}
}
