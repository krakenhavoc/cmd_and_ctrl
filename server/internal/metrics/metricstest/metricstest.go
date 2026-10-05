// Package metricstest reads metrics.Registry in tests: the value of a
// family summed over the series whose labels match, so a test can
// compare before and after an event. The event metrics are
// package-level and shared by every test in a binary, so a test counts
// a delta, never an absolute. Names lists what a set of collectors
// can report, for the check that the dashboards and alert rules name
// only real metrics.
package metricstest

import (
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

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

// fqName pulls the metric name out of a Desc's String(), which is the
// only place client_golang exposes it: Desc{fqName: "name", help: …}.
var fqName = regexp.MustCompile(`^Desc\{fqName: "([^"]+)"`)

// Names returns every metric family name the collectors describe. It
// reads Describe, not Gather, so a vector with no child yet (a counter
// nothing has incremented) is still named. It fails the test on a
// Desc that carries an error or whose name it cannot read.
func Names(t testing.TB, cs ...prometheus.Collector) map[string]bool {
	t.Helper()
	ch := make(chan *prometheus.Desc)
	go func() {
		for _, c := range cs {
			c.Describe(ch)
		}
		close(ch)
	}()
	out := map[string]bool{}
	for d := range ch {
		if err := d.Err(); err != nil {
			t.Errorf("collector describes an invalid metric: %v", err)
			continue
		}
		m := fqName.FindStringSubmatch(d.String())
		if m == nil {
			t.Errorf("cannot read a metric name from %s", d)
			continue
		}
		out[m[1]] = true
	}
	return out
}
