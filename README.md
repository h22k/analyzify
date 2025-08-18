## Analyzify Task

### Overview

This repository contains a Go-based event tracking system using ClickHouse as the database,
with a GraphQL API for recording and retrieving events.
All the information you need to get started, including setup and usage instructions, is available in this README.md.

---

### Task

- [x] Set up a Go project with appropriate module dependencies.
- [x] Use Docker to provide a local instance of ClickHouse for testing and development.

### 2. Schema Design

- [x] Define an event schema using Entgo. An event should have the following fields:
    - `EventID` (UUID)
    - `UserID` (UUID)
    - `EventType` (string)
    - `Timestamp` (datetime)
    - `Metadata` (JSONB or similar for additional event data)

### 3. API Endpoints

- [x] Implement GraphQL mutations to:
    - [x] Record a new event.
- [x] Implement GraphQL queries to:
    - [x] Retrieve events for a specific user.
    - [x] Aggregate and return event counts by `EventType` within a specified time range.

### 4. Database Integration

- [x] Configure Entgo to generate and manage the necessary schema in ClickHouse.
- [x] Implement repository functions to handle CRUD operations for the events.

--- 

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
  readability. Embedded SQL queries in Go variables via go:
  embed. [See the query file](./internal/db/clickhouse/sql/queries.go).
- Used a bit tricky way to implement EntGo with Clickhouse, because EntGo does not support Clickhouse
  natively. We had to use a custom connection object and some workarounds to make it
  work. [See the connection implementation](./internal/db/clickhouse/clickhouse.go#L39).

---

### What Could Have Been Done Better

The project follows best practices, focusing on simplicity, separation of concerns, and core functionality, guided by
the [Pareto Principle](https://en.wikipedia.org/wiki/Pareto_principle). Still, there are a few areas that could use some
improvement if this were a production-ready system.

- The error handling could have been more robust, especially in the GraphQL resolvers and database operations.
- The project could benefit from a clearer separation of concerns between the service and the database layer. For
  example, introducing a more distinct service layer could make future database changes easier and the codebase more
  maintainable.
- Separating environment variables into dedicated configuration files (e.g., .env.local, .env.prod) could further
  enhance maintainability and make the setup clearer and more readable.
- The project could benefit from implementing graceful shutdown for the HTTP server and database connections. Properly closing connections and releasing resources would make the application more robust and production-ready.
  - Pairing graceful shutdown with a bulk insert strategy for ClickHouse could make the system feel even more solid and reliable, ensuring smooth data insertion and efficient retrieval while everything shuts down cleanly.

---

### How to Run Locally

- Kept the setup and testing super simple so you can get the project running with just one command. Ensure that Docker
  and Make are installed and running on your system.

#### Setup

Once you run the command, the project will be accessible at [localhost:8080](http://localhost:8080), and the ClickHouse database will be ready at [localhost:9000](http://localhost:9000).

```bash
make dev
```

#### Testing

Since the project has a CI pipeline that runs tests on every push, you actually don't need to run tests locally.
[See the last action run](https://github.com/h22k/analyzify/actions/runs/17044904404/job/48318038389)
However, if you want to run the tests locally, you can do so with the following command.

```bash
make test
```