// Shared test constants. Uses the same default admin token as
// `server/Makefile`'s `dev` target ("dev-admin-token-not-for-production")
// so the test suite runs against a `make server-dev` stack without
// extra config.

export const ADMIN_TOKEN = "dev-admin-token-not-for-production";
