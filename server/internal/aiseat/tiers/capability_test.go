package tiers_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
)

// capability_test.go is #1060's regression gate, and the absence of
// it WAS the bug.
//
// #687's threat ordering and #1013's fuel pricing were both built,
// both tested, and both dead on every seat the lobby could create,
// because the only tests that asked "is this a TargetOrderer" asked a
// bare heuristic.New() — a policy no seat is ever given. The shipped
// tiers hand the runner a rules.Filter or a model.Policy, neither of
// which forwarded the hook, and a type assertion that answers false
// looks exactly like a policy with no opinion.
//
// So the question this file asks is the only one that would have
// caught it: take the policy the FACTORY builds for a tier — the
// object the lobby actually seats — and ask it everything the runner
// and the enumerator will ask it.

// capabilitiesOf is every optional Policy extension the runner or the
// enumerator looks for, answered against one built policy exactly the
// way the runner answers it.
//
// Keep this list in step with the assertions in aiseat: a new
// optional interface that nobody adds here is a new interface that
// can go dead on every tier without a test noticing, which is the
// whole of #1060.
type capabilities struct {
	targetOrderer bool // concede.go — legal.Options.OrderTargets (#687)
	fuelPricer    bool // concede.go — legal.Options.OrderCostFuel (#1013)
	conceder      bool // concede.go — the seat scoops
	tracer        bool // runner.go — DecideTraced for the decision log
	improviser    bool // improvise.go — ADR 0033 §8 (#686)
}

func capabilitiesOf(p aiseat.Policy) capabilities {
	var c capabilities
	_, c.targetOrderer = aiseat.Capability[aiseat.TargetOrderer](p)
	_, c.fuelPricer = aiseat.Capability[aiseat.CostFuelPricer](p)
	_, c.conceder = aiseat.Capability[aiseat.Conceder](p)
	_, c.tracer = aiseat.Capability[aiseat.Tracer](p)
	_, c.improviser = aiseat.Capability[aiseat.Improviser](p)
	return c
}

// Every shipped tier, built through the factory, offers the runner
// and the enumerator exactly the hooks its layers implement.
//
// `random` is the deliberate `false` column and is here for that
// reason: it has no opinion about a board, and a table that made it
// order targets would have narrowed the fuzzer for nothing. Every
// other tier has the heuristic somewhere underneath it, and the
// ordering is the heuristic's.
func TestEveryShippedTierForwardsItsOptionalHooks(t *testing.T) {
	f := tiers.NewFactory(tiers.FactoryOptions{Client: model.AlwaysIndex(0)})

	want := map[aiseat.Tier]capabilities{
		aiseat.TierRandom: {},
		aiseat.TierHeuristic: {
			targetOrderer: true, fuelPricer: true, conceder: true, tracer: true,
		},
		aiseat.TierAssisted: {
			targetOrderer: true, fuelPricer: true, conceder: true, tracer: true,
			improviser: true,
		},
		aiseat.TierStrong: {
			targetOrderer: true, fuelPricer: true, conceder: true, tracer: true,
			improviser: true,
		},
	}

	for _, tier := range tiers.All() {
		t.Run(string(tier), func(t *testing.T) {
			p, err := f.NewPolicy(botSeat(string(tier), ""))
			if err != nil {
				t.Fatalf("NewPolicy(%s): %v", tier, err)
			}
			got := capabilitiesOf(p)
			exp := want[aiseat.Tier(tier)]
			if got != exp {
				t.Errorf("tier %s, as the lobby builds it (%T), offers %+v; want %+v\n"+
					"a hook that answers false here is a feature that does not happen in "+
					"production, silently, on every table — see #1060", tier, p, got, exp)
			}
		})
	}
}

// The model tiers keep their hooks when the deployment has no model
// transport.
//
// A seat with no client is a complete Layer A + B policy under the
// tier's name (tiers.Options says so, and the outage drill depends on
// it), so it must still order its targets and price its fuel — those
// are the HEURISTIC's opinions and the heuristic is exactly what is
// left. What it loses is the model: no improvisation, and nothing to
// spend.
func TestModelTiersKeepTheHeuristicsHooksWithNoTransport(t *testing.T) {
	for _, tier := range []tiers.Tier{tiers.Assisted, tiers.Strong} {
		t.Run(string(tier), func(t *testing.T) {
			p, err := tiers.New(tier, tiers.Options{})
			if err != nil {
				t.Fatalf("New(%s): %v", tier, err)
			}
			got := capabilitiesOf(p)
			// Improviser and Spender are implemented by *model.Policy
			// itself, so they are present as INTERFACES even with no
			// client; what the missing transport removes is the
			// behaviour behind them, which improvise.go and Spend()
			// each answer for themselves.
			if !got.targetOrderer || !got.fuelPricer {
				t.Errorf("%s with no transport lost the heuristic's ordering hooks: %+v\n"+
					"the policy underneath a clientless funnel IS the heuristic, and #687 "+
					"and #1013 are its opinions", tier, got)
			}
		})
	}
}

// The chain is walked outward-in, so a wrapper that implements an
// extension ITSELF overrides the one underneath it rather than being
// bypassed by it.
//
// This is the property that makes the mechanism safe to apply to
// every optional interface at once. rules.Filter is a Tracer and its
// DecideTraced reports Layer A's verdict; if Capability preferred the
// innermost implementer, every heuristic-tier window would be
// reported as Layer B and the absorption rate ADR 0033 §5 rests on
// would come out of the decision log wrong.
func TestTheOutermostImplementerWins(t *testing.T) {
	p, err := tiers.New(tiers.Heuristic, tiers.Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	tr, ok := aiseat.Capability[aiseat.Tracer](p)
	if !ok {
		t.Fatal("the heuristic tier is not a Tracer")
	}
	if any(tr) != any(p) {
		t.Errorf("Capability[Tracer] returned %T, not the filter %T that wraps it — "+
			"Layer A's own verdict would never be reported", tr, p)
	}
}
