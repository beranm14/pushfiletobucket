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
	client, _ := storage.NewClient(ctx)
	bkt := client.Bucket(os.Getenv("BUCKET_NAME"))
	body, _ := ioutil.ReadAll(req.Body)
	sha256 := sha256.Sum256(body)
	obj := bkt.Object(string(sha256[:]))
	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, string(body[:])); err != nil {
		// TODO: Handle error.
	}
	// Close, just like writing a file.
	if err := w.Close(); err != nil {
		// TODO: Handle error.
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
