package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deeproot Elite — Creature — Merfolk Warrior {1}{G}, 1/1:
//
//	"Whenever another Merfolk you control enters, put a +1/+1 counter on
//	 target Merfolk you control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so
// it may be the Merfolk that just entered, the Elite itself or any other.
// If the target has gone by resolution nothing happens (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "764e8e6e-0ace-4fc8-8ca5-468a10ae3b73",
		Name:         "Deeproot Elite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Merfolk"))
				}, "Deeproot Elite — put a +1/+1 counter on target Merfolk you control",
					b36CounterOnChosenAnimal),
				TargetCreature("target Merfolk you control", YouControl(), HasSubtype("Merfolk")),
			),
		},
	})
}
