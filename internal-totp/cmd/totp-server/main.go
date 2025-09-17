package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"

	"internal-totp/internal/config"
	"internal-totp/internal/db"
	"internal-totp/internal/handlers"
	"internal-totp/internal/middleware"
)

func main() {
	cfg := config.Load()

	// init DB (open, pool, automigrate)
	if err := db.Init(cfg.DSN); err != nil {
		log.Fatalf("open mysql error: %v", err)
	}

	// http mux
	mux := http.NewServeMux()
	mux.HandleFunc("/enroll", middleware.WithAudit("enroll", handlers.HandleEnroll(cfg)))
	mux.HandleFunc("/activate", middleware.WithAudit("activate", handlers.HandleActivate))
	mux.HandleFunc("/verify", middleware.WithAudit("verify", handlers.HandleVerify))
	mux.HandleFunc("/disable", middleware.WithAudit("disable", handlers.HandleDisable))

	// CORS (方案B：rs/cors)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://192.168.151.64:5173", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	//handler := middleware.VaryOrigin(c.Handler(mux))

	log.Printf("TOTP service listening on %s", cfg.ListenAddr)
	log.Fatal(http.ListenAndServe(cfg.ListenAddr, c.Handler(mux)))
}
