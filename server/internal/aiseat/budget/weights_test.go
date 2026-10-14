package budget

import (
	"errors"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// weights_test.go is #3476: the cap weighs tokens by model and class,
// and a table has a share of the pool.

func TestWeightsPriceEachClassByModel(t *testing.T) {
	w := DefaultWeights()
	u := model.Usage{InputTokens: 1000, OutputTokens: 100, CacheReadTokens: 10_000, CacheWriteTokens: 800}
	// $/MTok x tokens = micro-dollars: haiku 1/5/0.1/1.25.
	if got, want := w.For("claude-haiku-4-5").Weigh(u), int64(1000+500+1000+1000); got != want {
		t.Errorf("haiku = %d, want %d", got, want)
	}
	// opus 5/25/0.5/6.25.
	if got, want := w.For("claude-opus-5").Weigh(u), int64(5000+2500+5000+5000); got != want {
		t.Errorf("opus = %d, want %d", got, want)
	}
	if h, o := w.For("CLAUDE-HAIKU-9").Weigh(u), w.For("claude-opus-5").Weigh(u); h >= o {
		t.Errorf("haiku (%d) should weigh less than opus (%d), whatever the case of the id", h, o)
	}
	// An id it does not know is priced as the dearest family.
	if w.For("mystery-model") != priceOpus || w.For("") != priceOpus {
		t.Errorf("unknown models must be priced as opus")
	}
	// A cache read is far cheaper than an output token.
	if w.For("claude-sonnet-4-5").Weigh(model.Usage{CacheReadTokens: 1}) >= w.For("claude-sonnet-4-5").Weigh(model.Usage{OutputTokens: 1}) {
		t.Errorf("cache read must weigh less than output")
	}
}

func TestPricesEnvOverridesAndFailsLoudly(t *testing.T) {
	w, err := weightsFrom(env(map[string]string{EnvPrices: "haiku=2/10/0.2/2.5, gpt-5=1.25/10/0.125/1.25"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := w.For("claude-haiku-4-5"); got.Input != 2 || got.Output != 10 {
		t.Errorf("override lost: %+v", got)
	}
	if got := w.For("gpt-5-mini"); got.Input != 1.25 {
		t.Errorf("extra family lost: %+v", got)
	}
	if got := w.For("claude-opus-5"); got != priceOpus {
		t.Errorf("a family not named kept its default: %+v", got)
	}
	for _, bad := range []string{"haiku", "haiku=1/2/3", "haiku=a/b/c/d", "haiku=-1/1/1/1", "=1/1/1/1"} {
		if _, err := weightsFrom(env(map[string]string{EnvPrices: bad})); err == nil {
			t.Errorf("%q: want an error", bad)
		}
	}
}

func TestBudgetCountsByTheModelThatAnswered(t *testing.T) {
	b := New(Limits{}, nil, WithLogger(quiet()))
	reply := func(modelID string) model.Client {
		return &model.FakeClient{Label: "x", Reply: func(int, model.Request) (model.Response, error) {
			return model.Response{Model: modelID, Usage: model.Usage{InputTokens: 1000, OutputTokens: 1000}}, nil
		}}
	}
	if err := call(t, b.Wrap(reply("claude-haiku-4-5"))); err != nil {
		t.Fatal(err)
	}
	if got := b.Status().DailyTokens; got != 6000 {
		t.Fatalf("after a haiku call = %d, want 6000", got)
	}
	if err := call(t, b.Wrap(reply("claude-opus-5"))); err != nil {
		t.Fatal(err)
	}
	if got := b.Status().DailyTokens; got != 6000+30000 {
		t.Fatalf("after an opus call = %d, want 36000", got)
	}
}

func TestTableShareLimitsOneTableNotTheOthers(t *testing.T) {
	// 1000 a day, 25% a table: 250 each.
	lim := Limits{DailyTokens: 1000, TableSharePercent: 25}
	if lim.TableLimit() != 250 {
		t.Fatalf("TableLimit = %d", lim.TableLimit())
	}
	b := New(lim, nil, WithLogger(quiet()), WithWeights(FlatWeights()))
	inner := usageClient(100, 0)
	a, c := b.ForTable(b.Wrap(inner), "table-a"), b.ForTable(b.Wrap(inner), "table-b")

	for i := 0; i < 3; i++ {
		if err := call(t, a); err != nil {
			t.Fatalf("table-a call %d: %v", i, err)
		}
	}
	// 300 >= 250: table-a is at its share and falls back.
	if err := call(t, a); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("table-a past its share: err = %v, want ErrBudgetExceeded", err)
	}
	// The pool is nowhere near its cap and other tables still get models.
	if err := call(t, c); err != nil {
		t.Fatalf("table-b: %v", err)
	}
	if st := b.Status(); st.State != StateOK || st.TableLimit != 250 || st.TablesAtShare != 1 || b.Blocked() != "" {
		t.Fatalf("status = %+v: one table at its share must not block the picker", st)
	}
	// A refused call counted nothing.
	if got := b.Status().DailyTokens; got != 400 {
		t.Fatalf("daily = %d, want 400", got)
	}
}

func TestTableShareResetsWithTheDay(t *testing.T) {
	clk := newClock("2026-10-12T23:59:00Z")
	b := New(Limits{DailyTokens: 1000, TableSharePercent: 10}, nil, WithClock(clk.now), WithLogger(quiet()), WithWeights(FlatWeights()))
	c := b.ForTable(b.Wrap(usageClient(100, 0)), "t")
	if err := call(t, c); err != nil {
		t.Fatal(err)
	}
	if err := call(t, c); !errors.Is(err, model.ErrBudgetExceeded) {
		t.Fatalf("at the share: %v", err)
	}
	clk.set("2026-10-13T00:01:00Z")
	if err := call(t, c); err != nil {
		t.Fatalf("new day: %v", err)
	}
}

func TestTableShareOffCases(t *testing.T) {
	for name, lim := range map[string]Limits{
		"share 0":      {DailyTokens: 100, TableSharePercent: 0},
		"share 100":    {DailyTokens: 100, TableSharePercent: 100},
		"no daily cap": {DailyTokens: 0, TableSharePercent: 10},
	} {
		if lim.TableLimit() != 0 {
			t.Errorf("%s: TableLimit = %d, want 0", name, lim.TableLimit())
		}
	}
	b := New(Limits{DailyTokens: 100, TableSharePercent: 100}, nil, WithLogger(quiet()), WithWeights(FlatWeights()))
	c := b.ForTable(b.Wrap(usageClient(40, 0)), "t")
	for i := 0; i < 2; i++ {
		if err := call(t, c); err != nil {
			t.Fatalf("call %d under the pool: %v", i, err)
		}
	}
}

// A client Wrap did not make (an own-key seat's) is left alone: it is
// not charged to a table and does not touch the pool's counters.
func TestForTableLeavesOtherClientsAlone(t *testing.T) {
	b := New(Limits{DailyTokens: 10, TableSharePercent: 10}, nil, WithLogger(quiet()), WithWeights(FlatWeights()))
	own := usageClient(1000, 1000)
	if got := b.ForTable(own, "t"); got != model.Client(own) {
		t.Fatalf("ForTable rebound a client that is not the pool's")
	}
	for i := 0; i < 3; i++ {
		if err := call(t, b.ForTable(own, "t")); err != nil {
			t.Fatalf("own-key style client refused by the pool: %v", err)
		}
	}
	if st := b.Status(); st.DailyTokens != 0 {
		t.Fatalf("daily = %d: an own-key client's spend is not the server's", st.DailyTokens)
	}
}

func TestTableShareEnv(t *testing.T) {
	got, err := limitsFrom(env(nil))
	if err != nil || got.TableSharePercent != DefaultTableSharePercent {
		t.Fatalf("default = %+v, %v", got, err)
	}
	got, err = limitsFrom(env(map[string]string{EnvTableShare: "40"}))
	if err != nil || got.TableSharePercent != 40 {
		t.Fatalf("40 = %+v, %v", got, err)
	}
	for _, bad := range []string{"-1", "101", "x", "2.5"} {
		_, err := limitsFrom(env(map[string]string{EnvTableShare: bad}))
		if err == nil || !strings.Contains(err.Error(), EnvTableShare) {
			t.Errorf("%q: err = %v", bad, err)
		}
	}
}

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
