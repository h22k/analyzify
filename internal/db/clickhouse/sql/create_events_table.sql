CREATE TABLE IF NOT EXISTS analyzify.events
(
    eventID
    UUID
    NOT
    NULL,
    userID
    UUID
    NOT
    NULL,
    event_type
    String
    NOT
    NULL,
    timestamp
    DateTime
    NOT
    NULL,
    metadata
    JSON,
    PRIMARY
    KEY
(
    userID,
    timestamp
)
    ) ENGINE = MergeTree
(
)
    ORDER BY
(
    userID,
    timestamp
);