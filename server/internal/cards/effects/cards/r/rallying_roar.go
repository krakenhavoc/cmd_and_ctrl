package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rallying Roar — Instant {2}{W} (Ixalan):
//
//	"Creatures you control get +1/+1 until end of turn. Untap them."
//
// "Them" is the creatures the first sentence affected: the set is fixed as
// the spell resolves (CR 611.2c), so the pump and the untap hit the same
// creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40ee79dd-cf53-415c-8e3c-d1a125a735db",
		Name:         "Rallying Roar",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Rallying Roar — creatures you control get +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
		},
	})
}
