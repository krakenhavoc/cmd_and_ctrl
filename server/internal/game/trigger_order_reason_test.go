package game

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// trigger_order_reason_test.go — ADR 0018's #3013 amendment, sub-PR 1:
// every trigger-order prompt is counted once, as it is queued, by why
// the engine asked. The reason never changes whether it asks.

func triggerOrderPromptCount(t *testing.T, reason TriggerOrderReason) float64 {
	t.Helper()
	return metricstest.Value(t, "cmdctrl_trigger_order_prompts_total", map[string]string{"reason": string(reason)})
}

// The engine's reasons and the metric's closed label set are the same
// list, other aside: a reason added on one side only would be counted
// as "other", or would be a label nothing reports.
func TestTriggerOrderReasonsAreMetricLabels(t *testing.T) {
	labels := metrics.TriggerOrderReasonLabels()
	var want []string
	for _, r := range TriggerOrderReasons() {
		want = append(want, string(r))
		if !slices.Contains(labels, string(r)) {
			t.Errorf("reason %q is not in metrics' reason labels", r)
		}
	}
	for _, l := range labels {
		if l != metrics.TriggerOrderReasonOther && !slices.Contains(want, l) {
			t.Errorf("metrics' reason label %q is no game.TriggerOrderReason", l)
		}
	}
}

// A prompt is counted once, under its reason, however many times the
// held drain runs before it is answered; a skipped batch is not
// counted.
func TestTriggerOrderPromptIsCountedOnceByReason(t *testing.T) {
	for _, c := range []struct {
		mode   TriggerOrderMode
		reason TriggerOrderReason
	}{
		{TriggerOrderWhenItMatters, TriggerOrderReasonUndescribed},
		{TriggerOrderAlways, TriggerOrderReasonAlways},
	} {
		g := newActiveGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, c.mode); err != nil {
			t.Fatal(err)
		}
		var log []string
		queueTwoDifferentForTest(g, me, &log)
		before := triggerOrderPromptCount(t, c.reason)
		for i := 0; i < 3; i++ {
			if drainForTest(g) {
				t.Fatalf("mode %q: the batch drained, want it held", c.mode)
			}
		}
		if findTriggerOrderPrompt(g, me.ID) == nil {
			t.Fatalf("mode %q: no prompt", c.mode)
		}
		if got := triggerOrderPromptCount(t, c.reason) - before; got != 1 {
			t.Errorf("mode %q: %q moved by %v over three held drains, want 1", c.mode, c.reason, got)
		}
	}

	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.SetTriggerOrderPreference(me.ID, TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	var log []string
	queueTwoDifferentForTest(g, me, &log)
	before := metricstest.Value(t, "cmdctrl_trigger_order_prompts_total", nil)
	if !drainForTest(g) {
		t.Fatal("never: the batch was held")
	}
	if got := metricstest.Value(t, "cmdctrl_trigger_order_prompts_total", nil) - before; got != 0 {
		t.Errorf("a skipped batch moved the counter by %v", got)
	}
}

// The reason agrees with the yes/no answer it rides on, in every mode.
func TestTriggerOrderAskReasonMatchesTheAnswer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var log []string
	a, b := queueTwoDifferentForTest(g, me, &log)
	for _, mode := range []TriggerOrderMode{TriggerOrderWhenItMatters, TriggerOrderAlways, TriggerOrderNever, "some-future-mode"} {
		for _, items := range [][]*StackItem{{a}, {a, b}, {a, a}} {
			var ask bool
			var reason TriggerOrderReason
			g.WithWriteLock(func() { ask, reason = g.triggerOrderAskLocked(items, mode) })
			if ask != (reason != "") {
				t.Errorf("mode %q, %d items: ask=%v with reason %q", mode, len(items), ask, reason)
			}
			var old bool
			g.WithWriteLock(func() { old = g.seatNeedsTriggerOrderLocked(items, mode) })
			if old != ask {
				t.Errorf("mode %q, %d items: seatNeedsTriggerOrderLocked %v, triggerOrderAskLocked %v", mode, len(items), old, ask)
			}
		}
	}
}
