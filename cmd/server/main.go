package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/h22k/analyzify/config"
	"github.com/h22k/analyzify/graph"
	"github.com/h22k/analyzify/internal/db/clickhouse"
	"github.com/h22k/analyzify/internal/service"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	cfg := config.Load()
	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	conn := clickhouse.NewConn(cfg.DbHost(), cfg.DbUser(), cfg.DbPass(), cfg.DbPort(), cfg.DbName())

	if err := conn.Migrate(); err != nil {
		cancel()
		if closeErr := conn.Close(); closeErr != nil {
			panic("failed to close ClickHouse connection: " + closeErr.Error())
		}
		panic("failed to migrate ClickHouse database: " + err.Error())
	}

	go func(port string) {
		srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
			EventService: service.NewEventService(conn),
		}}))

		srv.AddTransport(transport.Options{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})

		srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New[string](100),
		})

		http.Handle("/", playground.Handler("Analyzify playground", "/query"))
		http.Handle("/query", srv)

		log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}(cfg.AppPort())

	<-rootCtx.Done()

}
