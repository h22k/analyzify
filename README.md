### Architectural Decision Records (ADR)

This repository contains the Architectural Decision Records (ADRs) for the project. ADRs are a way to capture important
architectural decisions made during the development of a software project.

- Picked _userID_ and _timestamp_ as the primary key for the `events` table due to gain performance and simplicity. It
  is
  also recommended by Clickhouse
  documentation. [See the documentation](https://clickhouse.com/docs/guides/creating-tables#a-brief-intro-to-primary-keys).
- Used Nested data type for the `events` table to store multiple values in a single column (metadata). This is a
  Clickhouse-specific
  feature that allows for more efficient storage and querying of related
  data. [See the documentation](https://clickhouse.com/docs/sql-reference/data-types/nested-data-structures/nested).
    - This decision was made to optimize the storage and retrieval of event metadata, which can vary in structure and
      size. However, EntGo does not support
      Nested data type, so we had to use a JSON type for the `events` table for metadata column.
- Separated SQL queries into a separate file to improve maintainability and
  readability. Embedded SQL queries in Go via go:embed. [See the query file](./internal/db/clickhouse/sql/queries.go).
- Used a bit tricky way to implement EntGo with Clickhouse, because EntGo does not support Clickhouse
  natively. We had to use a custom connection object and some workarounds to make it
  work. [See the connection implementation](./internal/db/clickhouse/clickhouse.go#L39).