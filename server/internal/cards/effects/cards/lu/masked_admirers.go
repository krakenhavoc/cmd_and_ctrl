package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Masked Admirers — Creature — Elf Shaman {2}{G}{G}, 3/2 (Kaldheim
// Commander, #5506):
//
//	"When this creature enters, draw a card.
//	 Whenever you cast a creature spell, you may pay {G}{G}. If you do,
//	 return this card from your graveyard to your hand."
//
// The second ability works from the graveyard only (Silvan Reveler's shape):
// it triggers when its controller casts a creature spell, asks the payment as
// it resolves, and returns the very card that triggered, not a card of the
// same name that left and came back (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea97fc76-3a32-414c-aa12-87be1a0eb9f3",
		Name:         "Masked Admirers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Masked Admirers — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.InGraveyard(effects.WheneverYouCast(effects.Creature(), "Masked Admirers — you may pay {G}{G} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnToHand(g, item, "{G}{G}", "Masked Admirers — pay {G}{G} to return it to your hand?")
				})),
		},
	})
}
