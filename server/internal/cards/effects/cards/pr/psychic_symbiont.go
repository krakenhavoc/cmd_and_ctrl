package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psychic Symbiont — Creature — Nightmare Horror {4}{U}{B}, 3/3 (Commander
// Legends):
//
//	"Flying
//	 When this creature enters, target opponent discards a card and you
//	 draw a card."
//
// Flying is a printed keyword. The trigger targets an opponent; the
// opponent chooses the card they discard (CR 701.9b) and you draw one. If
// the target is illegal on resolution the whole ability is removed (CR
// 608.2b), so you draw nothing either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34439fb2-0272-4624-aa4c-9ea439b60489",
		Name:            "Psychic Symbiont",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Psychic Symbiont — target opponent discards a card and you draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						var opp uuid.UUID
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								opp = t.ID
							}
						}
						if opp == uuid.Nil {
							return nil
						}
						if err := effects.DiscardOne(g, item, opp); err != nil {
							return err
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
