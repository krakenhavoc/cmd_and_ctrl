package main

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// registerGameMetrics registers the games, players and WebSocket state
// collectors (ADR 0123 §3) on reg, over the live lobby and hub. With a
// database it also registers the users and database collectors; with
// none (CMDCTRL_DATA_DIR empty) those families are absent rather than
// zero. userStore is the users store main built: the users collector
// needs its SQL form, and only a database gives one.
func registerGameMetrics(reg prometheus.Registerer, log *slog.Logger, l *lobby.Lobby, hub *ws.Hub, database *db.DB, userStore users.Store) {
	reg.MustRegister(metrics.NewTablesCollector(l, hub))
	if database == nil {
		return
	}
	reg.MustRegister(metrics.NewDBCollector(database, log))
	if sqlUsers, ok := userStore.(*users.SQLStore); ok {
		reg.MustRegister(metrics.NewUsersCollector(sqlUsers, l, log))
	}
}
