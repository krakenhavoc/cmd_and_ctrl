package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rabid Rats — Creature — Rat {1}{B}, 1/1 (Stronghold, S82 OLD b750, #5627):
//
//	"{T}: Target blocking creature gets -1/-1 until end of turn."
//
// A tap ability on a target that is blocking as the target is chosen and again
// as the ability resolves (CR 608.2b), then BoostUntilEOT at -1/-1 on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f2504b0-bd5a-4b61-b8bb-c8b20aa1e26e",
		Name:         "Rabid Rats",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target blocking creature gets -1/-1 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Rabid Rats — -1/-1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
