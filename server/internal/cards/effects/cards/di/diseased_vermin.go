package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diseased Vermin — Creature — Rat {2}{B} 1/1 (Alliances, S82 OLD batch 818):
//
//	"Whenever this creature deals combat damage to a player, put an
//	 infection counter on it.
//	 At the beginning of your upkeep, this creature deals X damage to target
//	 opponent previously dealt damage by it, where X is the number of
//	 infection counters on it."
//
// The players it has dealt damage to are a per-player note on the object
// (Card.PlayerNotes, #6097): it is recorded when the combat damage trigger
// resolves, goes with the object when it leaves (CR 400.7), and the upkeep
// trigger's clause offers only noted opponents.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d18bea4-c9b2-4fc7-948f-2583589994d7",
		Name:         "Diseased Vermin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Diseased Vermin — put an infection counter on it", effects.OldB818VerminInfects),
			{
				Watches:     []game.EventKind{game.EventBeginUpkeep},
				AppliesTo:   effects.ByYou,
				TargetsFrom: effects.OldB818PreviouslyDamagedOpponent,
				Key:         "Diseased Vermin — X damage to target opponent previously dealt damage by it",
				Effect:      effects.OldB818VerminDealsInfectionDamage,
			},
		},
	})
}
