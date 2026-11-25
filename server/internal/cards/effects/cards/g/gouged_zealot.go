package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gouged Zealot — Creature — Cyclops Berserker {3}{R}, 4/3 (Modern Horizons
// 2, #5501):
//
//	"Reach
//	 Delirium — Whenever this creature attacks, if there are four or more
//	 card types among cards in your graveyard, this creature deals 1 damage
//	 to each creature defending player controls."
//
// An intervening "if" (CR 603.4): delirium is checked as it attacks and
// again as the trigger resolves. The creatures are those the defending
// player controls as it resolves, damaged together.
//
// No simplification.
func init() {
	zealot := effects.InterveningIf(effects.B24CondDelirium{}, effects.On(game.EventAttack, effects.ThisAttacked,
		"Gouged Zealot — 1 damage to each creature defending player controls (delirium)",
		effects.S82OldB38DamageEachCreatureDefendingPlayerControls(1)))
	zealot.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}}
	effects.Register(effects.Spec{
		OracleID:        "35b30274-bf8c-4d3c-b738-f9e62cdd5156",
		Name:            "Gouged Zealot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{zealot},
	})
}
