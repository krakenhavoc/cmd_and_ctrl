package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fourth Bridge Prowler — Creature — Human Rogue {B}, 1/1 (Aether Revolt, S82
// older sets, #5536):
//
//	"When this creature enters, you may have target creature get -1/-1 until
//	 end of turn."
//
// Any creature may be chosen, the Prowler itself included; the "may" is
// answered as the trigger resolves, and a target that is gone does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da18d97c-85f3-4395-9af5-bca72e33ebfc",
		Name:         "Fourth Bridge Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Fourth Bridge Prowler — you may have target creature get -1/-1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Fourth Bridge Prowler — -1/-1"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
				"Fourth Bridge Prowler — have target creature get -1/-1 until end of turn?"),
		},
	})
}
