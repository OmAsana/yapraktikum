package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OmAsana/yapraktikum/internal/server"
)

func logIncoming(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)

}

func startHTTPServer(wg *sync.WaitGroup, handler http.Handler) *http.Server {
	srv := &http.Server{Addr: "127.0.0.1:8080", Handler: handler}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println("Server shut down with err: ", err)
		}
	}()
	return srv

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	repo := server.NewRepositoryMock()
	metricsServer := server.NewMetricsServer(repo)
	//http.Handle(
	//	"/update/gauge/",
	//	pkg.CheckRequestMethod(metricsServer.UpdateGauge(), http.MethodPost),
	//)
	//http.Handle(
	//	"/update/counter/",
	//	pkg.CheckRequestMethod(metricsServer.UpdateCounters(), http.MethodPost),
	//)
	//http.HandleFunc("/update/", func(writer http.ResponseWriter, request *http.Request) {
	//	http.Error(writer, "not implemented", http.StatusNotImplemented)
	//
	//})
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	http.Error(w, "Not Found", http.StatusNotFound)
	//})

	waitServerShudown := &sync.WaitGroup{}

	waitServerShudown.Add(1)
	httpServer := startHTTPServer(waitServerShudown, metricsServer)

	go func() {
		<-ctx.Done()
		httpServer.Shutdown(ctx)
	}()

	waitServerShudown.Wait()

}
