package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grief — Creature — Elemental Incarnation {2}{B}{B}, 3/2:
//
//	"Menace
//	 When this creature enters, target opponent reveals their hand.
//	 You choose a nonland card from it. That player discards that card.
//	 Evoke—Exile a black card from your hand."
//
// The black incarnation: Solitude's evoke (EvokePitch, a black card
// from hand rather than mana, with the CR 702.74a sacrifice trigger
// bundled) over Thoughtseize's pick (ADR 0116) as an enters trigger.
// The pick's chooser is the trigger's controller (CR 113.8) and the
// opponent is the trigger's target. Grief can't pitch itself: it is
// already on the stack when the cost is paid (CR 601.2a).
//
// Evoked, the pick resolves before the sacrifice trigger, as Fury's and
// Solitude's do (the 2021-06-18 ruling allows that order). The whole table sees the revealed hand
// (CR 701.20a); a hand with no nonland card is revealed and nothing is
// discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d6b89dc-23fb-48e9-a54b-cea2838ae7c8",
		Name:            "Grief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{
			effects.EvokePitch(
				effects.CardInYourHand("a black card from your hand", effects.OfColor("B")),
				"a black card from your hand",
			),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Grief — target opponent reveals their hand, you choose a nonland card",
					effects.TargetRevealsYouChooseDiscardAbility(effects.Nonland(), "nonland card")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
