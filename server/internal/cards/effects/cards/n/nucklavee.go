package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nucklavee — Creature — Beast {4}{U/R}{U/R}, 4/4 (Commander 2011, #5574):
//
//	"When this creature enters, you may return target red sorcery card from
//	 your graveyard to your hand.
//	 When this creature enters, you may return target blue instant card from
//	 your graveyard to your hand."
//
// Two separate triggers, each aimed as it goes on the stack: a trigger with no
// card in your graveyard to name is not put there (CR 603.3d), and one whose
// card has left by resolution does nothing (CR 608.2b). The "you may" is the
// CR 603.5 prompt, asked after the target is chosen.
//
// No simplification.
func init() {
	returns := func(kind string, match effects.CardPredicate) game.TriggeredAbility {
		return effects.Targeting(effects.Optional(effects.WhenThisEnters("Nucklavee — return target "+kind+" card from your graveyard to your hand", effects.ReturnTargetCardToHand),
			"Nucklavee — return target "+kind+" card from your graveyard to your hand?"),
			effects.TargetCardInGraveyard("target "+kind+" card from your graveyard", match, effects.YouOwn()))
	}
	effects.Register(effects.Spec{
		OracleID:     "09c2d504-f2e1-4799-b948-c7b9502d5088",
		Name:         "Nucklavee",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			returns("red sorcery", effects.And(effects.OfColor("R"), effects.Sorcery())),
			returns("blue instant", effects.And(effects.OfColor("U"), effects.Instant())),
		},
	})
}
