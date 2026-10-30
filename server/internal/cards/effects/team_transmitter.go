package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Team Transmitter — {3} Artifact (Marvel Super Heroes Commander, #4455):
//
//	"Whenever a Hero you control enters, you gain 1 life.
//	 {T}: Add one mana of any color."
//
// The trigger sees every Hero that enters under your control while the
// Transmitter is on the battlefield; a Hero token counts like a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23714aed-0b26-4e21-b290-36c21815471c",
		Name:         "Team Transmitter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && hasSubtype(c, source.TextSubtype("Hero"))
			}, "Team Transmitter — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
	})
}
