package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleeting Memories — Enchantment {2}{U} (Shadows over Innistrad, #5543):
//
//	"When this enchantment enters, investigate.
//	 Whenever you sacrifice a Clue, target player mills three cards."
//
// "You sacrifice a Clue" is Tireless Tracker's reading (the event is
// emitted before the Clue leaves). The target is any player, chosen as the
// trigger goes on the stack and still legal as it resolves (CR 608.2b);
// the Clue sacrificed to Fleeting Memories' own investigate counts as any
// other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbb0173d-7ec9-4f43-b457-f8de6703497e",
		Name:         "Fleeting Memories",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Fleeting Memories — investigate", effects.Do(effects.Investigate{N: 1})),
			effects.Targeting(effects.On(game.EventSacrifice, effects.LonisYouSacrificedAClue,
				"Fleeting Memories — target player mills three cards", effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					player, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.MillCards{Player: player, N: 3}.Apply(ctx)
				}))),
				effects.TargetPlayer("target player")),
		},
	})
}
