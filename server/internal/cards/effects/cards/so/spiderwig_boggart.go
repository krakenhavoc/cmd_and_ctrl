package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiderwig Boggart — Creature — Goblin Shaman {2}{B}, 2/2 (Lorwyn):
//
//	"When this creature enters, target creature gains fear until end of turn."
//
// An enters trigger that targets any creature; fear lasts to the cleanup
// step, and a target that is gone or illegal on resolution is left alone
// (CR 608.2b). Fear is the keyword the block check honours (CR 702.36a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd0bbf91-1826-42c7-82af-6820646fa003",
		Name:         "Spiderwig Boggart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Spiderwig Boggart — target creature gains fear until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"fear"}, Label: "Spiderwig Boggart — fear"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
