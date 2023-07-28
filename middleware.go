package main

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
)

// カンマで区切られたオリジンをパースして配列に整形
func parseOrigins(originsENV string) []string {
	var origins []string
	if origins_ := strings.Split(originsENV, ","); len(origins_) != 0 {
		for _, origin := range origins_ {
			// 空白をTrim
			var origin = strings.ReplaceAll(origin, " ", "")
			if origin != "" {
				origins = append(origins, origin)
			}
		}
	}
	return origins
}

func setCors() cors.Config {
	return cors.Config{
		// アクセスを許可したいアクセス元
		AllowOrigins: parseOrigins(os.Getenv("ALLOW_ORIGIN")),
		// アクセスを許可したいHTTPメソッド(以下の例だとPUTやDELETEはアクセスできません)
		AllowMethods: []string{
			"POST",
			"GET",
			"OPTIONS",
		},
		// 許可したいHTTPリクエストヘッダ
		AllowHeaders: []string{
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		// cookieなどの情報を必要とするかどうか
		AllowCredentials: true,
		// preflightリクエストの結果をキャッシュする時間
		MaxAge: 24 * time.Hour,
	}
}
