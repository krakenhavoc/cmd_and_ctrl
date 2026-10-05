// Package metricstest reads metrics.Registry in tests: the value of a
// family summed over the series whose labels match, so a test can
// compare before and after an event. The event metrics are
// package-level and shared by every test in a binary, so a test counts
// a delta, never an absolute.
package metricstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// Value sums the family name over every series carrying all of labels
// (a subset match; nil matches every series): a counter's or a gauge's
// value, a histogram's sample count.
func Value(t testing.TB, name string, labels map[string]string) float64 {
	t.Helper()
	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	var sum float64
	for _, f := range families {
		if f.GetName() != name {
			continue
		}
		for _, m := range f.GetMetric() {
			have := map[string]string{}
			for _, lp := range m.GetLabel() {
				have[lp.GetName()] = lp.GetValue()
			}
			match := true
			for k, v := range labels {
				if have[k] != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
			switch {
			case m.GetCounter() != nil:
				sum += m.GetCounter().GetValue()
			case m.GetGauge() != nil:
				sum += m.GetGauge().GetValue()
			case m.GetHistogram() != nil:
				sum += float64(m.GetHistogram().GetSampleCount())
			}
		}
	}
	return sum
}

// L is shorthand for a label map.
type L = map[string]string
