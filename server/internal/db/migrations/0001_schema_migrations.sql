-- Migration 0001: the migrations ledger itself.
--
-- Every migration — including this one — is recorded here once it
-- commits, forward-only (ADR 0051 decision 1). This file only creates
-- the table; migrate() inserts this migration's own row in the same
-- transaction that runs the statement below.
CREATE TABLE schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at INTEGER NOT NULL
);
