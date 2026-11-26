package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oracle of Bones — {2}{R}{R} Creature — Minotaur Shaman 3/1 (Born of the Gods):
//
//	"Haste
//	 Tribute 2 (As this creature enters, an opponent of your choice may put two
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, you may cast an instant or
//	 sorcery spell from your hand without paying its mana cost."
//
// The offer is Wildfire Eternal's: asked as the trigger resolves, nothing asked of a
// hand with no instant or sorcery.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0503fc7e-6819-4979-b0df-7de23de4558b",
		Name:            "Oracle of Bones",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements:    effects.Tribute("Oracle of Bones", 2),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Oracle of Bones — you may cast an instant or sorcery spell from your hand without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					effects.OfferFreeHandCast(effects.NewContext(g, item), item.Controller, oracleOfBonesFreeCast,
						"an instant or sorcery spell", effects.FreeInstantOrSorcery())
					return nil
				}),
		},
	})
}

var oracleOfBonesFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/oracle-of-bones-cast", "Oracle of Bones")
