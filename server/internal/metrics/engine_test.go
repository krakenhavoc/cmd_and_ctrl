package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeLag RestorePointLag

func (f fakeLag) RestorePointLag(time.Time) RestorePointLag { return RestorePointLag(f) }

// Every engine and bot metric, exercised with every value its callers
// can produce and some they should not, still passes the label guard:
// an unregistered action type, an unknown seat kind, tier, layer, cause
// or result each lands on a value inside its closed set.
func TestEngineAndBotLabelsAreClosedSets(t *testing.T) {
	RegisterActionTypes("cast_spell", "pass_priority")
	for _, typ := range []string{"cast_spell", "pass_priority", ActionTypeBundle, ActionTypeLobby, "never_registered", ""} {
		for _, kind := range []string{SeatKindHuman, SeatKindBot, SeatKindAgent, SeatKindAdmin, "spectator"} {
			RecordAction(typ, kind, true, time.Millisecond)
			RecordAction(typ, kind, false, 0)
		}
	}
	EffectError()
	SetBootRestore(2, 1, 1, 3)

	for _, tier := range []string{BotTierRandom, BotTierHeuristic, BotTierAssisted, BotTierStrong, "scripted", "rules+heuristic"} {
		for _, layer := range []string{BotLayerA, BotLayerB, BotLayerC, BotLayerRandom, "", "stub-b"} {
			RecordBotDecision(tier, layer)
		}
		ObserveBotDecision(tier, time.Second)
		for _, c := range append(append([]string{}, runnerFallbackCauses...), "something-new") {
			RecordRunnerFallback(tier, c)
		}
		for c := range modelFallbackCauses {
			RecordModelFallback(tier, c)
		}
		RecordModelFallback(tier, "something-new")
		for _, r := range append(append([]string{}, modelResultLabels...), "weird") {
			RecordBotModelCall(tier, r, time.Second)
		}
		AddBotModelTokens(tier, 10, 2)
	}

	reg := NewRegistry()
	reg.MustRegister(NewRestorePointCollector(fakeLag{Oldest: 90 * time.Second, Behind: 2}))
	if err := CheckClosedLabels(reg); err != nil {
		t.Fatal(err)
	}
}

// An action type nobody registered is "other", never the client's
// string. bundle and lobby pass through: ws sends them for its own
// commits only, and maps a client's type to the enum or "other" before
// it gets here.
func TestActionTypeLabel(t *testing.T) {
	RegisterActionTypes("tap")
	cases := map[string]string{
		"tap":              "tap",
		"untap_everything": ActionTypeOther,
		"":                 ActionTypeOther,
		ActionTypeBundle:   ActionTypeBundle,
		ActionTypeLobby:    ActionTypeLobby,
	}
	for in, want := range cases {
		if got := actionTypeLabel(in); got != want {
			t.Errorf("actionTypeLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

// One commit is one count and one observation, on the child its labels
// name.
func TestRecordActionCountsOnce(t *testing.T) {
	RegisterActionTypes("mulligan")
	c := actionsTotal.WithLabelValues("mulligan", SeatKindAgent, ResultRejected)
	before := testutil.ToFloat64(c)
	beforeObs := applyHistogramCount(t, ResultRejected)
	RecordAction("mulligan", SeatKindAgent, false, 2*time.Millisecond)
	if got := testutil.ToFloat64(c) - before; got != 1 {
		t.Errorf("cmdctrl_actions_total rose by %v, want 1", got)
	}
	if got := applyHistogramCount(t, ResultRejected) - beforeObs; got != 1 {
		t.Errorf("cmdctrl_action_apply_seconds{result=rejected} observed %d, want 1", got)
	}
}

func applyHistogramCount(t *testing.T, result string) uint64 {
	t.Helper()
	reg := prometheus.NewRegistry()
	reg.MustRegister(actionApplySeconds)
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range families {
		for _, m := range f.GetMetric() {
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "result" && lp.GetValue() == result {
					return m.GetHistogram().GetSampleCount()
				}
			}
		}
	}
	return 0
}

// The collector reports what the reader says, in seconds.
func TestRestorePointCollector(t *testing.T) {
	c := NewRestorePointCollector(fakeLag{Oldest: 90 * time.Second, Behind: 2})
	want := `
# HELP cmdctrl_rooms_behind_restore_point Active rooms whose last applied commit is not yet captured in a restore point.
# TYPE cmdctrl_rooms_behind_restore_point gauge
cmdctrl_rooms_behind_restore_point 2
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want), "cmdctrl_rooms_behind_restore_point"); err != nil {
		t.Error(err)
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)
	families, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range families {
		if f.GetName() == "cmdctrl_restore_point_age_seconds" {
			if got := f.GetMetric()[0].GetGauge().GetValue(); got != 90 {
				t.Errorf("cmdctrl_restore_point_age_seconds = %v, want 90", got)
			}
			return
		}
	}
	t.Error("cmdctrl_restore_point_age_seconds was not collected")
}

// The boot gauges hold what the restore pass set, and a later pass
// with nothing to restore sets them back to zero rather than leaving
// the old numbers.
func TestSetBootRestore(t *testing.T) {
	SetBootRestore(3, 1, 2, 4)
	if got := testutil.ToFloat64(bootRestoreGames.WithLabelValues(RestoreOutcomeAbandoned)); got != 2 {
		t.Errorf("abandoned = %v, want 2", got)
	}
	if got := testutil.ToFloat64(bootRestoreDegradedCards); got != 4 {
		t.Errorf("degraded cards = %v, want 4", got)
	}
	SetBootRestore(0, 0, 0, 0)
	for _, o := range restoreOutcomeLabels {
		if got := testutil.ToFloat64(bootRestoreGames.WithLabelValues(o)); got != 0 {
			t.Errorf("%s = %v after an empty pass, want 0", o, got)
		}
	}
}

// The cached child is the child: a cache hit increments the same
// series a fresh lookup would.
func TestActionCounterCacheHitsTheSameSeries(t *testing.T) {
	RegisterActionTypes("untap")
	k := actionKey{typ: "untap", kind: SeatKindBot, applied: true}
	first := actionCounter(k)
	second := actionCounter(k)
	if first != second {
		t.Fatal("two lookups of one key returned different counters")
	}
	before := testutil.ToFloat64(actionsTotal.WithLabelValues("untap", SeatKindBot, ResultApplied))
	second.Inc()
	if got := testutil.ToFloat64(actionsTotal.WithLabelValues("untap", SeatKindBot, ResultApplied)) - before; got != 1 {
		t.Errorf("the cached counter is not the series: rose by %v", got)
	}
}

// A commit after the first of its kind allocates nothing for its
// labels (ADR 0123 asks the hot path to stay cheap).
func TestRecordActionDoesNotAllocate(t *testing.T) {
	RegisterActionTypes("draw_card")
	RecordAction("draw_card", SeatKindHuman, true, time.Millisecond)
	if n := testing.AllocsPerRun(100, func() {
		RecordAction("draw_card", SeatKindHuman, true, time.Millisecond)
	}); n != 0 {
		t.Errorf("RecordAction allocates %v times per call, want 0", n)
	}
}
