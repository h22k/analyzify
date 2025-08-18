test:
	docker compose --profile test up -d clickhouse-test
	docker compose --profile test up --abort-on-container-exit --exit-code-from app-test app-test
	docker compose --profile test down --volumes
dev:
	docker compose --profile dev up --detach
	docker compose --profile dev logs -f