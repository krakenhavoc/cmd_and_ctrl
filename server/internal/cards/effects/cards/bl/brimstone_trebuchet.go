package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brimstone Trebuchet — Artifact Creature — Wall {2}{R}, 1/3 (Throne of
// Eldraine, #5516):
//
//	"Defender, reach
//	 {T}: This creature deals 1 damage to each opponent.
//	 Whenever a Knight you control enters, untap this creature."
//
// Defender and reach are printed keywords. The untap is Blasting Station's
// "untap this permanent" body; the Trebuchet is not a Knight, so only
// another permanent entering can trigger it.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8dd601f-3615-4cef-b837-52329cb361f8",
		Name:            "Brimstone Trebuchet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: This creature deals 1 damage to each opponent.",
			Cost:  effects.TapCost(),
			// Damage to each opponent answers nothing in response.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageToEachOpponent(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Knight"))
			}, "Brimstone Trebuchet — untap this creature", effects.B33UntapSelf),
		},
	})
}
