package main

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/budget"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// registerGameMetrics registers every state collector (ADR 0123 §3) on
// reg: the restore-point gauges over the room manager, and the games,
// players and WebSocket gauges over the live lobby and hub. With a
// database it also registers the users and database collectors; with
// none (CMDCTRL_DATA_DIR empty) those families are absent rather than
// zero. userStore is the users store main built: the users collector
// needs its SQL form, and only a database gives one.
//
// It is the one place main registers a collector beyond
// metrics.Collectors, so the dashboards' name check
// (TestMonitoringConfigNamesRegisteredMetrics) can see every name the
// server reports by calling it.
func registerGameMetrics(reg prometheus.Registerer, log *slog.Logger, mgr *ws.RoomManager, l *lobby.Lobby, hub *ws.Hub, database *db.DB, userStore users.Store) {
	// How far a deploy now would rewind the live tables.
	reg.MustRegister(metrics.NewRestorePointCollector(mgr))
	reg.MustRegister(metrics.NewTablesCollector(l, hub))
	// The caps and what is carried against them (ADR 0148).
	reg.MustRegister(metrics.NewCapacityCollector(l))
	if database == nil {
		return
	}
	reg.MustRegister(metrics.NewDBCollector(database, log))
	if sqlUsers, ok := userStore.(*users.SQLStore); ok {
		reg.MustRegister(metrics.NewUsersCollector(sqlUsers, l, log))
		reg.MustRegister(metrics.NewGrowthCollector(sqlUsers, log))
	}
}

// registerBotMetrics registers the bot watch gauges (docs/monitoring.md):
// the longest a bot seat has owed an action, always, and the hosted-model
// spend against its caps when there is a spend guard (b is nil on a
// server with no model transport, and the families are then absent
// rather than zero).
func registerBotMetrics(reg prometheus.Registerer, seats metrics.BotSeatSource, b lobby.BotModelBudget) {
	reg.MustRegister(metrics.NewBotSeatCollector(seats))
	if b != nil {
		reg.MustRegister(metrics.NewBudgetCollector(budgetSource{b}))
	}
}

// budgetSource reads the spend guard's Status for the budget collector.
type budgetSource struct{ b lobby.BotModelBudget }

func (s budgetSource) BudgetReading() metrics.BudgetReading {
	st := s.b.Status()
	return metrics.BudgetReading{
		DayTokens: st.DailyTokens, DayLimit: st.DailyLimit,
		MonthTokens: st.MonthlyTokens, MonthLimit: st.MonthlyLimit,
		Blocked: st.State != budget.StateOK,
	}
}
