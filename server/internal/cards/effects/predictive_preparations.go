package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Predictive Preparations — Sorcery {1}{W} (Reality Fracture, tracker #2795):
//
//	"Put a +1/+1 counter on each of one or two target creatures.
//	 Flashback {3}{W}"
//
// One clause counted one-to-two; the counters go on every target that is
// still legal as the spell resolves (CR 608.2b). Flashback is the shared
// constructor: the graveyard is opened by CastableZones and the card
// exiles itself on leaving the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d06a5642-861b-4b0d-9ea5-07046b7e1d37",
		Name:             "Predictive Preparations",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("one or two target creatures").WithCount(1, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
