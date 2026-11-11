package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kefnet the Mindful — Legendary Creature — God {2}{U}, 5/5 (Duskmourn
// Commander, #5368):
//
//	"Flying, indestructible
//	 Kefnet can't attack or block unless you have seven or more cards in
//	 hand.
//	 {3}{U}: Draw a card, then you may return a land you control to its
//	 owner's hand."
//
// The attack half is a clause of the creature's own attack restriction
// (game.AttackTargetRestriction.ControllerHandAtLeast, the city's
// blessing's shape), read live at declaration; the block half is Vantress
// Gargoyle's CR 509.1b block rule. The land is chosen after the draw and
// "you may" lets the controller keep it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "caf626a8-7abc-496b-9d3a-23194ce70e01",
		Name:            "Kefnet the Mindful",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
		Static:          []game.StaticAbility{CantAttackUnlessYouHaveCardsInHand(7)},
		BlockRules: []game.BlockRule{CantBlockUnless(YouHaveCardsInHandAtLeast(7),
			"it can't block unless its controller has seven or more cards in hand")},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Draw a card, then you may return a land you control to its owner's hand.",
			Cost:    ManaCost("{3}{U}"),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  cmdr2S22DrawThenMayReturnALand,
		}},
	})
}
