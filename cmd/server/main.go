// Command server 為本機可橋接的 WebSocket 對局伺服器入口。
//
// 它把已實作完成的權威 Hub（配對 + RuleEngine 驗證 + 廣播 + 斷線重連）掛上監聽埠：
// 兩個客戶端連上同一位址即被配成一局，走法在兩端之間轉送並廣播。橋接邏輯全在
// server.Hub，本檔僅負責「使其可執行」。
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/yuanyu90221/xiangqi-platform/server"
)

func main() {
	addr := flag.String("addr", ":8080", "WebSocket 伺服器監聽位址")
	flag.Parse()

	hub := server.NewHub()
	mux := http.NewServeMux()
	mux.Handle("/", hub.Handler())

	log.Printf("象棋對局伺服器監聽中：%s（ws://<host>%s）", *addr, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("伺服器結束：%v", err)
	}
}
