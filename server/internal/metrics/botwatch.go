package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The bot watch gauges, read at scrape time:
//
//	cmdctrl_bot_seat_owed_seconds_max  how long the longest-waiting bot seat has owed an action
//	cmdctrl_bot_budget_tokens{period}  hosted-model tokens used in the UTC day or month
//	cmdctrl_bot_budget_limit_tokens{period}  the cap for it; 0 means no cap
//	cmdctrl_bot_budget_blocked         1 while model calls are refused (a cap, or the kill switch)
//
// A bot seat "owes" an action from the moment it has a legal move to
// make until it makes one (or the engine shows it has none). A seat
// that owes for minutes is wedged: the table waits on it.

// Budget periods: the period label.
const (
	BudgetDay   = "day"
	BudgetMonth = "month"
)

var budgetPeriods = []string{BudgetDay, BudgetMonth}

// BotSeatSource reports the bot runners. *aiseat.Manager satisfies it.
type BotSeatSource interface {
	// MaxOwed is the longest any live bot seat has owed an action; zero
	// when none does.
	MaxOwed() time.Duration
}

// BudgetReading is the hosted-model spend guard's state.
type BudgetReading struct {
	DayTokens, DayLimit     int64
	MonthTokens, MonthLimit int64
	// Blocked is true while calls are refused: a cap reached, or the
	// admin kill switch.
	Blocked bool
}

// BudgetSource reads the spend guard. main.go adapts the budget to it.
type BudgetSource interface {
	BudgetReading() BudgetReading
}

var (
	botOwedDesc = prometheus.NewDesc("cmdctrl_bot_seat_owed_seconds_max",
		"Seconds the longest-waiting bot seat has owed an action it has not taken: it has a legal move, and nothing has been played. 0 when no bot seat is waiting.",
		nil, nil)
	budgetTokensDesc = prometheus.NewDesc("cmdctrl_bot_budget_tokens",
		"Hosted-model tokens counted against the spend cap in the current UTC day or month.",
		[]string{"period"}, nil)
	budgetLimitDesc = prometheus.NewDesc("cmdctrl_bot_budget_limit_tokens",
		"The hosted-model spend cap for a period, in tokens. 0 means no cap.",
		[]string{"period"}, nil)
	budgetBlockedDesc = prometheus.NewDesc("cmdctrl_bot_budget_blocked",
		"1 while model-backed bots are refused model calls (a cap reached or the admin kill switch), else 0.",
		nil, nil)
)

type botSeatCollector struct{ src BotSeatSource }

// NewBotSeatCollector returns the collector of
// cmdctrl_bot_seat_owed_seconds_max.
func NewBotSeatCollector(src BotSeatSource) prometheus.Collector {
	return &botSeatCollector{src: src}
}

func (c *botSeatCollector) Describe(ch chan<- *prometheus.Desc) { ch <- botOwedDesc }

func (c *botSeatCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(botOwedDesc, prometheus.GaugeValue, c.src.MaxOwed().Seconds())
}

type budgetCollector struct{ src BudgetSource }

// NewBudgetCollector returns the collector of the cmdctrl_bot_budget_*
// gauges. main.go registers it only when a spend guard exists.
func NewBudgetCollector(src BudgetSource) prometheus.Collector {
	return &budgetCollector{src: src}
}

func (c *budgetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- budgetTokensDesc
	ch <- budgetLimitDesc
	ch <- budgetBlockedDesc
}

func (c *budgetCollector) Collect(ch chan<- prometheus.Metric) {
	r := c.src.BudgetReading()
	ch <- prometheus.MustNewConstMetric(budgetTokensDesc, prometheus.GaugeValue, float64(r.DayTokens), BudgetDay)
	ch <- prometheus.MustNewConstMetric(budgetTokensDesc, prometheus.GaugeValue, float64(r.MonthTokens), BudgetMonth)
	ch <- prometheus.MustNewConstMetric(budgetLimitDesc, prometheus.GaugeValue, float64(r.DayLimit), BudgetDay)
	ch <- prometheus.MustNewConstMetric(budgetLimitDesc, prometheus.GaugeValue, float64(r.MonthLimit), BudgetMonth)
	blocked := 0.0
	if r.Blocked {
		blocked = 1
	}
	ch <- prometheus.MustNewConstMetric(budgetBlockedDesc, prometheus.GaugeValue, blocked)
}
