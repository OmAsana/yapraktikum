package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/OmAsana/yapraktikum/internal/pkg"
	"github.com/OmAsana/yapraktikum/internal/server"
)

func logIncoming(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)

}

func startHttpServer(wg *sync.WaitGroup) *http.Server {
	srv := &http.Server{Addr: "127.0.0.1:8080"}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server shut down with err: %v", err)
		}
	}()
	return srv

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	repo := server.NewRepositoryMock()
	metricsServer := server.NewMetricsServer(repo)
	http.Handle(
		"/update/gauge/",
		pkg.CheckRequestMethod(metricsServer.UpdateGauge(), http.MethodPost),
	)
	http.Handle(
		"/update/counter/",
		pkg.CheckRequestMethod(metricsServer.UpdateCounters(), http.MethodPost),
	)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	waitServerShudown := &sync.WaitGroup{}

	waitServerShudown.Add(1)
	httpServer := startHttpServer(waitServerShudown)

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx)
	}()

	go func() {
		debugTicker := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-debugTicker.C:
				repo.ListStoredMetrics()
			}
		}
	}()

	waitServerShudown.Wait()

}
