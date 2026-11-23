package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeproot Elite — Creature — Merfolk Warrior {1}{G}, 1/1:
//
//	"Whenever another Merfolk you control enters, put a +1/+1 counter on
//	 target Merfolk you control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so
// it may be the Merfolk that just entered, the Elite itself or any other.
// If the target has gone by resolution nothing happens (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "764e8e6e-0ace-4fc8-8ca5-468a10ae3b73",
		Name:         "Deeproot Elite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Merfolk"))
				}, "Deeproot Elite — put a +1/+1 counter on target Merfolk you control",
					effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target Merfolk you control", effects.YouControl(), effects.HasSubtype("Merfolk")),
			),
		},
	})
}
