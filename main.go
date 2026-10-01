// Command taskflow serve a API e o front do Taskflow.
//
// Uso:
//
//	go run .            # http://localhost:3000
//	PORT=8080 go run .  # porta alternativa
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"taskflow/internal/api"
	"taskflow/internal/store"
)

const (
	defaultPort = 3000
	maxAttempts = 10
	envDataFile = "TASKFLOW_DATA_FILE"
)

func main() {
	if err := run(); err != nil {
		slog.Error("taskflow encerrado com erro", "erro", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dataFile := store.ResolveFile(envDataFile, ".")
	dataStore, err := store.New(dataFile)
	if err != nil {
		return fmt.Errorf("preparar armazenamento em %s: %w", dataFile, err)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", preferredPort()),
		Handler:           api.New(dataStore, "public", logger).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		logger.Info("encerrando o Taskflow")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("falha no encerramento gracioso", "erro", err)
		}
	}()

	logger.Info("dados carregados", "arquivo", dataFile)
	return listenWithFallback(server, preferredPort(), os.Getenv("PORT") != "")
}

// listenWithFallback sobe o servidor na porta desejada. Se ela estiver
// ocupada, avança para a próxima até um limite razoável — a menos que a porta
// tenha sido escolhida explicitamente via PORT.
func listenWithFallback(server *http.Server, port int, explicit bool) error {
	attempts := maxAttempts
	for {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			slog.Info(fmt.Sprintf("TASKFLOW rodando em http://localhost:%d", port))
			if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		}
		if explicit || attempts <= 1 || !isAddressInUse(err) {
			return err
		}
		next := port + 1
		slog.Warn(fmt.Sprintf("porta %d ocupada, tentando a porta %d...", port, next))
		port = next
		attempts--
	}
}

func isAddressInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}

func preferredPort() int {
	custom := os.Getenv("PORT")
	if custom == "" {
		return defaultPort
	}
	port, err := strconv.Atoi(custom)
	if err != nil || port < 1 || port > 65535 {
		slog.Warn("PORT inválido, usando a porta padrão", "valor", custom, "porta", defaultPort)
		return defaultPort
	}
	return port
}
