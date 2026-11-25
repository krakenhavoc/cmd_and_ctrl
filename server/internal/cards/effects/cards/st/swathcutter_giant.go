package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swathcutter Giant — Creature — Giant Soldier {4}{R}{W}, 5/5 (Guilds of
// Ravnica, #5525):
//
//	"Vigilance
//	 Whenever this creature attacks, it deals 1 damage to each creature
//	 defending player controls."
//
// The defending player is the one attacked, or the controller of the
// planeswalker or battle attacked; the set of creatures is fixed as the
// ability resolves, and the damage is one instance.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventAttack, effects.ThisAttacked,
		"Swathcutter Giant — 1 damage to each creature defending player controls",
		effects.S82OldB38DamageEachCreatureDefendingPlayerControls(1))
	trigger.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}
	effects.Register(effects.Spec{
		OracleID:        "fc16a2ce-a570-454f-8498-f9c604c0d68b",
		Name:            "Swathcutter Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
