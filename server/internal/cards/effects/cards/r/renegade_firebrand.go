package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renegade Firebrand — Creature — Human Warrior {2}{R}, 3/2 (Kaladesh,
// #5539):
//
//	"As long as you control a Chandra planeswalker, this creature gets +1/+0
//	 and has first strike."
//
// Two statics on this creature alone (Goblin Blast-Runner's shape: a layer 7c
// pump and a layer 6 keyword) gated on Karplusan Hound's reading of "a Chandra
// planeswalker": a planeswalker with the Chandra subtype, not a card merely
// named Chandra. Both end the moment the last one leaves.
//
// No simplification.
func init() {
	chandra := func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Chandra") }
	strike := effects.KeywordGrant(effects.WhileYouHave(chandra, effects.SelfCreature), "first strike")
	strike.Label = "As long as you control a Chandra planeswalker, this creature has first strike."
	effects.Register(effects.Spec{
		OracleID:     "d4b0a6e6-937d-428a-a469-bc09c30ac276",
		Name:         "Renegade Firebrand",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control a Chandra planeswalker, this creature gets +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.WhileYouHave(chandra, effects.SelfCreature),
				Apply:     effects.PumpBy(1, 0),
			},
			strike,
		},
	})
}
