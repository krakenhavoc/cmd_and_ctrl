package game

import "github.com/google/uuid"

// closure_cleanup_probes_test.go — #4253's final cleanup deleted closure
// doors only tests still called (ExileSpellThenForEffect,
// DealDamageToPlayerThenForEffect, ChoosePlayerPrompt.Then, ColorPrompt.Then,
// PileSplitPrompt.Then). The tests that used them now ask through the keyed
// forms, and hear the continuation through these registered probes, which
// record what they were told in package-level state each test resets.

// tailHeard is one run of testProbeTail.
type tailHeard struct {
	IDs    []uuid.UUID
	Amount int
	Word   string
}

var (
	// testProbeHeard is every run of testProbeTail, in order.
	testProbeHeard []tailHeard
	// testProbeTail records each outcome it is told.
	testProbeTail = RegisterTail("test/4253-probe-tail", func(_ *Game, o TailOutcome) error {
		testProbeHeard = append(testProbeHeard, tailHeard{IDs: copyUUIDs(o.IDs), Amount: o.Amount, Word: o.Tail.Word})
		return nil
	})

	// testProbePicked is every answer testProbeOptionThen was handed.
	testProbePicked []OptionPicked
	// testProbeOptionThen records each option-pick answer it is handed.
	testProbeOptionThen = RegisterOptionPickThen("test/4253-probe-option", func(_ *Game, r OptionPicked) error {
		testProbePicked = append(testProbePicked, r)
		return nil
	})

	// testPlotIfExiled is "it becomes plotted" after an exile of a spell,
	// told the spell only when it reached exile.
	testPlotIfExiled = RegisterTail("test/4253-plot-if-exiled", func(g *Game, o TailOutcome) error {
		if len(o.IDs) == 1 {
			g.PlotExiledCardForEffect(o.IDs[0], uuid.Nil)
		}
		return nil
	})
)

// resetProbes clears what the probes have heard.
func resetProbes() {
	testProbeHeard, testProbePicked = nil, nil
}

// probeTail is a keyed tail on testProbeTail.
func probeTail() KeyedTail { return KeyedTail{Ref: testProbeTail} }

// probeExiled reports what testProbeTail heard as an exile's "exiled"
// answers: one bool per run, true when the run was told a card.
func probeExiled() []bool {
	out := make([]bool, 0, len(testProbeHeard))
	for _, h := range testProbeHeard {
		out = append(out, len(h.IDs) > 0)
	}
	return out
}

// probeDealt reports the damage each run of testProbeTail was told.
func probeDealt() []int {
	out := make([]int, 0, len(testProbeHeard))
	for _, h := range testProbeHeard {
		out = append(out, h.Amount)
	}
	return out
}

// probeChosen reports the seat each answer testProbeOptionThen was handed
// names: uuid.Nil when nobody was chosen.
func probeChosen() []uuid.UUID {
	out := make([]uuid.UUID, 0, len(testProbePicked))
	for _, r := range testProbePicked {
		seat := uuid.Nil
		if r.Option != nil {
			seat = r.Option.Player
		}
		out = append(out, seat)
	}
	return out
}

// probePiles is the last pile split testProbeTail heard: the pile taken
// and the pile left (a pile tail is told both as its IDs, taken first,
// with how many were taken as its Amount). ran is how many runs it heard.
func probePiles() (taken, left []uuid.UUID, ran int) {
	if len(testProbeHeard) == 0 {
		return nil, nil, 0
	}
	h := testProbeHeard[len(testProbeHeard)-1]
	n := min(max(h.Amount, 0), len(h.IDs))
	return h.IDs[:n], h.IDs[n:], len(testProbeHeard)
}

var (
	// testEarthbendPower is the power of the tail's carried land when
	// testEarthbendTail ran (-1 when it was not on the battlefield), and
	// testEarthbendRuns how many times it ran.
	testEarthbendPower, testEarthbendRuns = -1, 0
	testEarthbendTail                     = RegisterTail("test/4253-earthbend-land-power", func(g *Game, o TailOutcome) error {
		testEarthbendRuns++
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.battlefieldCardLocked(o.Tail.Carry[0]); ok {
			testEarthbendPower = c.CurrentPower()
		}
		return nil
	})

	// testEnteredSeen is, for each run of testEnteredTail, whether the
	// permanent it was told about was on the battlefield as it ran.
	testEnteredSeen []bool
	testEnteredTail = RegisterTail("test/4253-entered-on-battlefield", func(g *Game, o TailOutcome) error {
		testProbeHeard = append(testProbeHeard, tailHeard{IDs: copyUUIDs(o.IDs)})
		on := false
		if len(o.IDs) == 1 {
			_, on = g.battlefieldCardLocked(o.IDs[0])
		}
		testEnteredSeen = append(testEnteredSeen, on)
		return nil
	})
)

// probeEntered is the permanent each run of testProbeTail (or
// testEnteredTail) was told entered: uuid.Nil when nothing did.
func probeEntered() []uuid.UUID {
	out := make([]uuid.UUID, 0, len(testProbeHeard))
	for _, h := range testProbeHeard {
		id := uuid.Nil
		if len(h.IDs) > 0 {
			id = h.IDs[0]
		}
		out = append(out, id)
	}
	return out
}
