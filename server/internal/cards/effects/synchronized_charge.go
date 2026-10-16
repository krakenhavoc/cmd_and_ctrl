package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Synchronized Charge — Sorcery {1}{G} (#3931):
//
//	"Distribute two +1/+1 counters among one or two target creatures you
//	 control. Creatures you control with counters on them gain vigilance
//	 and trample until end of turn.
//	 Harmonize {4}{G}"
//
// Abzan Charm's distribution (CR 601.2d: each target at least one, the
// split announced with the targets), then Mutational Advantage's set:
// the creatures you control that have any counter on them as the second
// sentence is applied, the two just placed included, locked as it
// resolves (CR 611.2c), so a creature that gets a counter later this turn
// does not gain them. Harmonize is CR 702.180.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "faa0055b-c1d2-4ce7-875f-41b4be8030bd",
		Name:             "Synchronized Charge",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{G}")},
		Targets: TargetCreature("one or two target creatures you control", YouControl()).
			WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl(), b64HasAnyCounter()),
				Keywords: []string{"vigilance", "trample"},
				Label:    "Synchronized Charge — vigilance and trample",
			}.Apply(ctx)
		},
	})
}
