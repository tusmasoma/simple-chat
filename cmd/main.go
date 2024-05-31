package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/tusmasoma/simple-chat/config"
	"github.com/tusmasoma/simple-chat/interface/handler"
	"github.com/tusmasoma/simple-chat/repository/redis"
	"github.com/tusmasoma/simple-chat/repository/sqlite"
	"github.com/tusmasoma/simple-chat/repository/websocket"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":8083", "tcp host:port to connect")
	flag.Parse()

	ctx := context.Background()

	srv := &http.Server{
		Addr:    addr,
		Handler: Init(ctx),
	}

	/* ===== サーバの起動 ===== */
	log.SetFlags(0)
	log.Println("Server running...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, os.Kill)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Server stopping...")

	tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(tctx); err != nil {
		log.Println("failed to shutdown http server", err)
	}
	log.Println("Server exited")
}

func Init(ctx context.Context) *chi.Mux {
	db := config.InitDB()
	defer db.Close()

	cacheClient := config.NewClient()

	userRepo := sqlite.NewUserRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)

	pubsubRepo := redis.NewPubSubRepository(cacheClient)

	hub := websocket.NewHubWebSocketRepository(ctx, roomRepo, userRepo, pubsubRepo)

	handler := handler.NewWebsocketHandler(hub, nil)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin"},
		ExposedHeaders:   []string{"Link", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	go hub.Run()

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.WebSocketConnection(w, r)
	})

	return r
}
