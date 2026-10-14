package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeBots time.Duration

func (f fakeBots) MaxOwed() time.Duration { return time.Duration(f) }

type fakeBudget BudgetReading

func (f fakeBudget) BudgetReading() BudgetReading { return BudgetReading(f) }

func TestBotSeatCollectorReportsSeconds(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewBotSeatCollector(fakeBots(150 * time.Second)))
	if err := testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP cmdctrl_bot_seat_owed_seconds_max Seconds the longest-waiting bot seat has owed an action it has not taken: it has a legal move, and nothing has been played. 0 when no bot seat is waiting.
# TYPE cmdctrl_bot_seat_owed_seconds_max gauge
cmdctrl_bot_seat_owed_seconds_max 150
`)); err != nil {
		t.Error(err)
	}
}

func TestBudgetCollectorReportsUseLimitAndBlock(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewBudgetCollector(fakeBudget{DayTokens: 900, DayLimit: 1000, MonthTokens: 5, MonthLimit: 0, Blocked: true}))
	n, err := testutil.GatherAndCount(reg)
	if err != nil || n != 5 {
		t.Fatalf("series = %d, err %v; want 5", n, err)
	}
	if err := testutil.GatherAndCompare(reg, strings.NewReader(`
# HELP cmdctrl_bot_budget_blocked 1 while model-backed bots are refused model calls (a cap reached or the admin kill switch), else 0.
# TYPE cmdctrl_bot_budget_blocked gauge
cmdctrl_bot_budget_blocked 1
`), "cmdctrl_bot_budget_blocked"); err != nil {
		t.Error(err)
	}
}

func TestAuthFailedCountsByReason(t *testing.T) {
	for _, reason := range authFailureReasons {
		before := testutil.ToFloat64(authFailures.WithLabelValues(reason))
		AuthFailed(reason)
		if got := testutil.ToFloat64(authFailures.WithLabelValues(reason)); got != before+1 {
			t.Errorf("%s: %v, want %v", reason, got, before+1)
		}
	}
}

func TestInstrumentHTTPCountsRateLimited(t *testing.T) {
	mux := NewServeMux()
	mux.HandleFunc("GET /ratelimited-probe", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	mux.HandleFunc("GET /fine-probe", func(http.ResponseWriter, *http.Request) {})
	h := InstrumentHTTP(mux)
	before := testutil.ToFloat64(httpRateLimited.WithLabelValues("GET /ratelimited-probe"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ratelimited-probe", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/fine-probe", nil))
	if got := testutil.ToFloat64(httpRateLimited.WithLabelValues("GET /ratelimited-probe")); got != before+1 {
		t.Errorf("429 count = %v, want %v", got, before+1)
	}
	if got := testutil.ToFloat64(httpRateLimited.WithLabelValues("GET /fine-probe")); got != 0 {
		t.Errorf("a 200 counted as rate limited: %v", got)
	}
}
