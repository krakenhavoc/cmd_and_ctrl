package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shambling Cie'th — Creature — Mutant Horror {2}{B}, 3/3 (Final
// Fantasy, #4462):
//
//	"This creature enters tapped.
//	 Whenever you cast a noncreature spell, you may pay {B}. If you do,
//	 return this card from your graveyard to your hand."
//
// The second ability works only from the graveyard (Silvan Reveler's
// InGraveyard), and asks for the payment as it resolves: the card
// returns only if {B} was really paid, and only the very object that
// triggered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3f4aa87-a89e-43f8-abc3-712e6bf08889",
		Name:         "Shambling Cie'th",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.WheneverYouCast(effects.Noncreature(), "Shambling Cie'th — you may pay {B} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayForGraveyardReturnToHand(g, item, "{B}", "Shambling Cie'th — pay {B} to return it to your hand?")
				})),
		},
	})
}
