package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brain Gorgers — Creature — Zombie {3}{B}, 4/2:
//
//	"When you cast this spell, any player may sacrifice a creature of
//	 their choice. If a player does, counter Brain Gorgers.
//	 Madness {1}{B}"
//
// ADR 0150 (#1903): the sacrifice is offered to every player in APNAP
// order from the active player, the caster included, until one makes it
// (anyPlayerMayCounterThisSpell). The payer picks the creature, and a
// player with no creature can't pay (CR 118.3). The offer holds the
// spell on the stack until it is answered. A madness cast is a cast, so
// it triggers too. Madness is the engine's (#657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b9369ae-8827-45d1-af85-c777cde8e1bc",
		Name:         "Brain Gorgers",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{B}",
		Triggered: []game.TriggeredAbility{
			effects.AnyPlayerMayCounterThisSpell("Brain Gorgers", "sacrifice a creature",
				effects.SacrificePayment(1, "creature", "creatures", effects.QueryType("creature")), effects.AnyPlayerCounterThisSpell),
		},
	})
}
