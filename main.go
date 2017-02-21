package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {

	// Start server
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuid := r.URL.Query().Get("uuid")
		if uuid != "" {
			sd := fmt.Sprintf(`
    	{
    		"alexa:all": {
    			"productID": "%s",
    			"productInstanceAttributes": {
    				"deviceSerialNumber": "%s"
    			}
    		}
    	}`, os.Getenv("PRODUCT_ID"), uuid)
			u, err := url.Parse("https://www.amazon.com/ap/oa")
			if err != nil {
				log.Fatal(err)
			}

			q := u.Query()
			q.Add("client_id", os.Getenv("CLIENT_ID"))
			q.Add("scope", "alexa:all")
			q.Add("scope_data", sd)
			q.Add("response_type", "code")
			q.Add("redirect_uri", fmt.Sprintf("https://%s", r.Host))

			u.RawQuery = q.Encode()

			w.Header().Add("Location", u.String())
			w.WriteHeader(302)
			return
		}

		code := r.URL.Query().Get("code")
		form := url.Values{}

		form.Add("client_id", os.Getenv("CLIENT_ID"))
		form.Add("client_secret", os.Getenv("CLIENT_SECRET"))
		form.Add("code", code)
		form.Add("grant_type", "authorization_code")
		form.Add("redirect_uri", fmt.Sprintf("https://%s", r.Host))

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

		var authDetails struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		}

		err = json.NewDecoder(resp.Body).Decode(&authDetails)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte(fmt.Sprintf(`
      <!DOCTYPE html>
      <html>
        <head>
        <meta charset="UTF-8">
        <title>HueAlexa Auth</title>
        </head>
        <body>
          <form target="%s">
            <input type="hidden" name="access_token" value="%s" />
            <input type="hidden" name="refresh_token" value="%s" />
            <input type="hidden" name="expires_in" value="%d" />
            <input type="submit" value="Auth"/>
          </form>
        </body>
      </html>
    `, "", authDetails.AccessToken, authDetails.RefreshToken, authDetails.ExpiresIn)))

	})); err != nil {
		panic(err)
	}
}
