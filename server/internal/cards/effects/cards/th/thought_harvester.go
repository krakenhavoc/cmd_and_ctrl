package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought Harvester — {3}{U} Creature — Eldrazi Drone 2/4 (Oath of the
// Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 Whenever you cast a colorless spell, target opponent exiles the top card
//	 of their library."
//
// Skittering Cicada's "whenever you cast a colorless spell" over a target
// opponent chosen as the trigger goes on the stack (CR 603.3d). The exile is
// one card off the top of that player's library, no more if it is empty.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13f9a845-d0e9-4790-99e0-b857204ad1da",
		Name:            "Thought Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Colorless(), "Thought Harvester — target opponent exiles the top card of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := effects.TargetedPlayerOf(ctx)
					if id == uuid.Nil {
						return nil
					}
					return effects.MillToZone{Player: id, N: 1, To: game.ZoneExile}.Apply(ctx)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
