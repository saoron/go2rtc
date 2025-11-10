package api

import (
	"net/http"
	"strings"

	"github.com/AlexxIT/go2rtc/internal/app"
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
		if strings.Contains(r.Host, "dride.cloud") && app.IsProtectedPath(r.URL.Path) && !app.VerifyAssetToken(token)  {
			http.Error(w, "403 - Forbidden", http.StatusForbidden)
			return
		}
		
		fileServer.ServeHTTP(w, r)
	})
}



