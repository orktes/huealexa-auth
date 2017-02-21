package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {

	// Start server
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		form := url.Values{}

		form.Add("client_id", os.Getenv("CLIENT_ID"))
		form.Add("client_secret", os.Getenv("CLIENT_SECRET"))
		form.Add("code", code)
		form.Add("grant_type", "authorization_code")
		form.Add("redirect_uri", fmt.Sprintf("https://%s%s", r.Host, r.RequestURI))

		req, err := http.NewRequest("POST", "https://api.amazon.com/auth/o2/token", strings.NewReader(form.Encode()))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)

		io.Copy(w, resp.Body)

	})); err != nil {
		panic(err)
	}
}
