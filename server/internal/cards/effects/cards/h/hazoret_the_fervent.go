package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hazoret the Fervent — {3}{R} Legendary Creature — God 5/4 (Amonkhet, S82 old
// batch 248):
//
//	"Indestructible, haste
//	 Hazoret can't attack or block unless you have one or fewer cards in hand.
//	 {2}{R}, Discard a card: Hazoret deals 2 damage to each opponent."
//
// The attack half is game.AttackTargetRestriction.ControllerHandAtMost, read
// live at declaration (CR 508.1c); the block half is CantBlockUnless over the
// same count, read as blockers are declared (CR 509.1b). The discard is paid
// as a cost (CR 602.2b); the damage is dealt by Hazoret, to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c6934b6-67c4-457a-ae8f-59914fd8f256",
		Name:            "Hazoret the Fervent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible", "haste"},
		Static:          []game.StaticAbility{effects.CantAttackUnlessYouHaveAtMostCardsInHand(1)},
		BlockRules: []game.BlockRule{effects.CantBlockUnless(effects.OldB248YouHaveAtMostCardsInHand(1),
			"it can't block unless its controller has one or fewer cards in hand")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Hazoret deals 2 damage to each opponent.",
			Purpose: game.Purpose{DamageEachOpponent: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 2}),
		}},
	})
}
