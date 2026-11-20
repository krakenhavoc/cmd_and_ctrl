package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur War Shaman — Creature — Human Shaman {4}{G}{G}, 4/5 (Duskmourn
// Commander):
//
//	"When this creature enters, manifest the top card of your library.
//	 Whenever a permanent you control is turned face up, if it's a
//	 creature, you may have it fight target creature you don't control."
//
// The permanent is face up and the same object by the time the event is
// emitted (CR 708.8), so "if it's a creature" reads its real front face;
// the condition is the trigger's own and is rechecked as it resolves
// (CR 603.4). The target is chosen as the trigger is put on the stack and
// the "may" is asked as it resolves; the turned-up creature, not the Shaman,
// does the fighting, and nothing happens if either fighter has left or the
// target is no longer a creature you don't control (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb683e18-50b8-4821-9a1a-86c6dea63ac9",
		Name:         "Temur War Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Temur War Shaman — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
			effects.Optional(effects.Targeting(effects.On(game.EventTurnedFaceUp, effects.ACreatureYouControlWasTurnedFaceUp,
				"Temur War Shaman — have it fight target creature you don't control",
				effects.Cmdr2S37TurnedUpCreatureFights),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
				"Temur War Shaman — have the creature fight target creature you don't control?"),
		},
	})
}
