package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Devourer — Creature — Demon Cleric {1}{B}, 0/3:
//
//	"Each nonland card in your hand without foretell has foretell. Its
//	 foretell cost is equal to its mana cost reduced by {2}. (During your
//	 turn, you may pay {2} and exile it from your hand face down. Cast it on
//	 a later turn for its foretell cost.)
//	 Whenever you foretell a card, this creature gets +2/+0 until end of
//	 turn."
//
// The first sentence is a special-action grant over the hand
// (effects.ForetellFromYourHand): a nonland card in your hand that prints no
// foretell is offered the ordinary foretell special action for {2}, and the
// foretell cost it is cast for later is its mana cost with {2} of its generic
// mana taken off, never below zero and never touching coloured mana. A card
// that prints foretell keeps its own cost, and with no Devourer the card has
// none. The trigger is the foretell special action itself, so every card
// foretold, printed or granted, pumps it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "0ff7440b-b831-40a9-81c4-99d77bf93b8d",
		Name:                "Dream Devourer",
		Completeness:        effects.CompletenessFull,
		SpecialActionGrants: []game.SpecialActionGrant{effects.ForetellFromYourHand(2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSpecialAction, effects.YouForetoldACard,
				"Dream Devourer — gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2,
						Label: "Dream Devourer — +2/+0 until end of turn"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
