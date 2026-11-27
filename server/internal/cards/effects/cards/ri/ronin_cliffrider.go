package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ronin Cliffrider — Creature — Human Samurai {3}{R}{R}, 2/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Bushido 1
//	 Whenever this creature attacks, you may have it deal 1 damage to each
//	 creature defending player controls."
//
// Bushido 1 is the shared keyword trigger (CR 702.45a). The attack trigger is
// Swathcutter Giant's, made optional: the creatures are those the defending
// player controls as it resolves, damaged together, by the Cliffrider (its
// last-known information if it has gone).
//
// No simplification.
func init() {
	trigger := effects.On(game.EventAttack, effects.ThisAttacked,
		"Ronin Cliffrider — you may have it deal 1 damage to each creature defending player controls",
		effects.S82OldB38DamageEachCreatureDefendingPlayerControls(1))
	trigger.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}
	effects.Register(effects.Spec{
		OracleID:     "ce8fc1bb-5dd0-41b1-909a-bbfe6b165158",
		Name:         "Ronin Cliffrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bushido(1, "Ronin Cliffrider"),
			effects.Optional(trigger, "Ronin Cliffrider — have it deal 1 damage to each creature defending player controls?"),
		},
	})
}
