package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seraph Sanctuary — Land (EDHREC rank 2257):
//
//	"When this land enters, you gain 1 life.
//	 Whenever an Angel you control enters, you gain 1 life.
//	 {T}: Add {C}."
//
// The Angel deck's colourless utility land. Two triggers on one
// land, each a real stack item: its own entry (b06SelfETB), and
// every Angel that enters under its controller's control —
// effective subtypes, so a changeling Angel counts. The mana ability
// is Buried Ruin's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b504dc6-61cc-4a72-907c-145fa4c72466",
		Name:         "Seraph Sanctuary",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seraph Sanctuary — you gain 1 life", Do(GainLife{Amount: 1})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21AngelYouControlEntered(ev, source, g)
			}, "Seraph Sanctuary — an Angel entered, you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
