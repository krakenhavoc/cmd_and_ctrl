package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmagorian — Creature — Horror {5}{B}{B}, 6/6:
//
//	"When you cast this spell, any player may discard three cards. If a
//	 player does, counter Phantasmagorian.
//	 Discard three cards: Return this card from your graveyard to your
//	 hand."
//
// ADR 0150 (#1903): the discard is offered to every player in APNAP
// order from the active player, the caster included, until one makes it
// (anyPlayerMayCounterThisSpell). The payer picks the three cards, and
// a player with fewer than three can't pay (CR 118.3). The offer holds
// the spell on the stack until it is answered.
//
// The second ability works from the graveyard (CR 113.6), with a
// discard cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "248e0dfa-0bcb-402c-badf-ca59f02792f7",
		Name:         "Phantasmagorian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AnyPlayerMayCounterThisSpell("Phantasmagorian", "discard three cards", effects.DiscardPayment(3), effects.AnyPlayerCounterThisSpell),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard three cards: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.DiscardN(3, "three cards"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToHand,
		}},
	})
}
