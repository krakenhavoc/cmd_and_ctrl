package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bohn, Beguiling Balladeer — Legendary Creature — Human Bard {1}{U}{R},
// 3/3 (OLD b028, #6153):
//
//	"Each nonland card in your hand without foretell has foretell. Its
//	 foretell cost is equal to its mana cost reduced by {2}. (During your
//	 turn, you may pay {2} and exile it from your hand face down. Cast it
//	 on a later turn for its foretell cost.)
//	 Whenever you cast your second spell each turn, goad target creature
//	 an opponent controls."
//
// The first sentence is Dream Devourer's special-action grant over the
// hand (ForetellFromYourHand, CR 702.143a): a nonland card in your hand
// that prints no foretell is offered the foretell special action for {2},
// and is cast later for its mana cost less {2} of generic mana; a card that
// prints foretell keeps its own cost. The trigger is Devoted Duelist's
// "second spell each turn" (YouCastYourSecondSpellEachTurn) with Bothersome
// Quasit's targeted goad (CR 701.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "ab31b652-ddf2-480f-a955-e8b5c87728f9",
		Name:                "Bohn, Beguiling Balladeer",
		Completeness:        effects.CompletenessFull,
		SpecialActionGrants: []game.SpecialActionGrant{effects.ForetellFromYourHand(2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
					"Bohn, Beguiling Balladeer — goad target creature an opponent controls", effects.GoadTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
