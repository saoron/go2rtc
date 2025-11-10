package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/AlexxIT/go2rtc/www"
)

func initStatic(staticDir string) {
	var root http.FileSystem
	
	if staticDir != "" {
		log.Info().Str("dir", staticDir).Msg("[api] serve static")
		root = http.Dir(staticDir)
	} else {
		root = http.FS(www.Static)
	}

	base := len(basePath)
	fileServer := http.FileServer(root)

	HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		if base > 0 {
			r.URL.Path = r.URL.Path[base:]
		}
		token := r.URL.Query().Get("token")
		if strings.Contains(r.Host, "dride.cloud") && !VerifyAssetToken(token) {
			http.Error(w, "403 - Forbidden", http.StatusForbidden)
			return
		}
		
		fileServer.ServeHTTP(w, r)
	})
}



type AssetToken struct {
	Token     string
	ExpiresAt int64
}

func VerifyAssetToken(tokenToCompare string) bool {
	if tokenToCompare == "" {
		return false
	}

	tokens := DBRead("assetTokens")
	if tokens == "" {
		tokens = "[]"
	}
	var tokensObject []AssetToken
	err := json.Unmarshal([]byte(tokens), &tokensObject)
	if err != nil {
		fmt.Println("Failed to unmarshal asset token " + err.Error())
		return false
	}

	for _, token := range tokensObject {
		fmt.Println(token.Token, token.ExpiresAt-time.Now().Unix())
		if (token.Token == tokenToCompare) && (token.ExpiresAt-time.Now().Unix()) > 0 {
			return true
		}
	}
	fmt.Println("token not found!")
	return false
}

func DBRead(key string) string {
	key = strings.ToLower(key)
	resp, err := Get("http://127.0.0.1:4000/read?key="+key, 1)

	if err != nil {
		return ""
	}
	return resp
}

func Get(url string, retry int) (string, error) {
	if retry <= 0 {
		return "", errors.New("Max retries reached " + url)
	}

	resp, err := http.Get(url)

	if err != nil {
		time.Sleep(time.Millisecond * 100)
		return Get(url, retry-1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 250)
		return Get(url, retry-1)
	}
	return string(body), nil
}