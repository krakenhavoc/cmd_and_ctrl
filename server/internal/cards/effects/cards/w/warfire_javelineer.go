package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warfire Javelineer — Creature — Minotaur Warrior {3}{R}, 2/3
// (Amonkhet, #5535):
//
//	"When this creature enters, it deals X damage to target creature an opponent controls, where X is the number of instant and sorcery cards in your graveyard."
//
// Cyclops Electromancer's enters trigger and body; X is read as
// the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a7a1e9-23b4-4cde-baf1-b1c0165c6729",
		Name:         "Warfire Javelineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Warfire Javelineer — deals X damage to target creature an opponent controls, X the number of instant and sorcery cards in your graveyard",
					effects.CyclopsElectromancerBlast),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
