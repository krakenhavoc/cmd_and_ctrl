package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niv-Mizzet, Visionary — Legendary Creature — Dragon Wizard {4}{U}{R},
// 5/5 (Foundations, #4465):
//
//	"Flying
//	 You have no maximum hand size.
//	 Whenever a source you control deals noncombat damage to an opponent,
//	 you draw that many cards."
//
// Thor, Guardian of Midgard's trigger with a draw in place of the exile:
// one trigger per damage event, so two sources are two triggers. "A
// source you control" is read off the damage event
// (b34DamageSourceController: a permanent, or the spell still on the
// stack), and "that many" is the event's amount. No maximum hand size is
// Spec.NoMaxHandSize (Curiosity Crafter's bit); flying rides
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9f1dc61-831e-4244-a2da-aa9da6fd5b50",
		Name:            "Niv-Mizzet, Visionary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		NoMaxHandSize:   true,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ASourceYouControlDealtNoncombatDamageToAnOpponent,
				"Niv-Mizzet, Visionary — draw that many cards",
				effects.Cmdr2B08DrawThatMany),
		},
	})
}
