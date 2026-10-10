package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Memorial Vault — Artifact {3}{R}:
//
//	"{T}, Sacrifice another artifact: Exile the top X cards of your
//	 library, where X is one plus the mana value of the sacrificed
//	 artifact. You may play those cards this turn."
//
// The sacrifice is a cost, so X is read from the sacrificed artifact's
// last-known mana value (CR 608.2h) off the payment record. The exile is
// the impulse-exile primitive with the "play" form (not cast-only): a
// land exiled this way can be played, subject to the usual land-drop and
// timing rules, and the permission lapses at end of turn. A token has
// mana value 0, so sacrificing a Treasure exiles one card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10dda10f-2549-47e5-9a8b-fb414cf38d3f",
		Name:         "Memorial Vault",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another artifact: Exile the top X cards of your library, where X is one plus the mana value of the sacrificed artifact. You may play those cards this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SacrificedPermanent()
				if !ok {
					return nil
				}
				return ExileTopWithPermission{
					From:    item.Controller,
					GrantTo: item.Controller,
					N:       1 + info.ManaValue,
				}.Apply(ctx)
			},
		}},
	})
}
