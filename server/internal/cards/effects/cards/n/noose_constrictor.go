package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noose Constrictor — Creature — Snake {1}{G}, 2/2 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Reach
//	 Discard a card: This creature gets +1/+1 until end of turn."
//
// The discard is a cost the activator names at announce. The pump is a
// Layer 7c effect on this creature, applied only while it is still on
// the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb1d521f-cb64-4b87-b3f4-a74e91a60349",
		Name:            "Noose Constrictor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.DiscardN(1, "a card"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				src := ctx.Source()
				if !effects.OnBattlefield(g, src) {
					return nil
				}
				return effects.BoostUntilEOT{Target: src, Power: 1, Toughness: 1, Label: "Noose Constrictor — +1/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
