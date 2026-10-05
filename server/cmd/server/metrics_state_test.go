package main

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func familyNames(t *testing.T, g prometheus.Gatherer) map[string]bool {
	t.Helper()
	families, err := g.Gather()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, f := range families {
		out[f.GetName()] = true
	}
	return out
}

// ADR 0123 §3: the tables collector is always registered; the users and
// database families only with a database. With none they are absent,
// not zero.
func TestRegisterGameMetrics(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	dbFamilies := []string{"cmdctrl_users", "cmdctrl_users_played", "cmdctrl_db_size_bytes", "cmdctrl_db_backup_last_success_timestamp_seconds"}

	// No database.
	mgr := ws.NewRoomManager(log, "")
	reg := prometheus.NewRegistry()
	registerGameMetrics(reg, log, mgr, lobby.NewLobby(mgr), ws.NewHub(log), nil, users.NoStore{})
	names := familyNames(t, reg)
	for _, want := range []string{"cmdctrl_games", "cmdctrl_seats", "cmdctrl_seats_connected", "cmdctrl_ws_connections"} {
		if !names[want] {
			t.Errorf("no database: %s missing", want)
		}
	}
	for _, absent := range dbFamilies {
		if names[absent] {
			t.Errorf("no database: %s reported", absent)
		}
	}

	// A database.
	dir := t.TempDir()
	database, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	mgr = ws.NewRoomManager(log, dir)
	reg = prometheus.NewRegistry()
	registerGameMetrics(reg, log, mgr, lobby.NewLobbyWithStore(mgr, lobby.NewSQLStore(database)), ws.NewHub(log),
		database, users.NewSQLStore(database, nil))
	names = familyNames(t, reg)
	for _, want := range dbFamilies {
		if !names[want] {
			t.Errorf("with a database: %s missing", want)
		}
	}
	if err := metrics.CheckClosedLabels(reg); err != nil {
		t.Fatalf("label rule broken:\n%v", err)
	}
}
