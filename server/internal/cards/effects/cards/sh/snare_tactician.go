package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snare Tactician — Creature — Human Soldier {2}{W}, 2/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Whenever you cycle a card, tap target creature an opponent controls."
//
// Prickly Marmoset's watcher (every cycling of a card by the controller,
// CR 702.29b) with a target chosen as the trigger goes on the stack; with
// no creature an opponent controls it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efee89a6-b730-45b6-bb84-33de0f25ad47",
		Name:         "Snare Tactician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCycle("Snare Tactician — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
