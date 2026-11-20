package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drumhunter — Creature — Human Druid Warrior {3}{G}, 2/2 (Shards of
// Alara, #5588):
//
//	"At the beginning of your end step, if you control a creature with
//	 power 5 or greater, you may draw a card.
//	 {T}: Add {C}."
//
// The end-step trigger is an intervening "if" (CR 603.4), through the
// shared alaB1AtYourEndStepIfPowerFive row: it triggers only with a
// creature of power 5 or more on its controller's side, and is removed
// without a question if none is left as it resolves. The draw is a "you
// may" (CR 603.5). The mana ability adds one colourless.
//
// No simplification.
func init() {
	const label = "Drumhunter — you may draw a card"
	Register(Spec{
		OracleID:     "3aab84f2-731c-4079-9623-0b0c433442c3",
		Name:         "Drumhunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(alaB1AtYourEndStepIfPowerFive(label, Do(DrawCards{N: 1})),
				"Drumhunter — draw a card?", 1),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
