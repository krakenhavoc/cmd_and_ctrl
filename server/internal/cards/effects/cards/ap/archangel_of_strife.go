package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archangel of Strife — Creature — Angel {5}{W}{W}, 6/6 (Commander 2011, #5574):
//
//	"Flying
//	 As this creature enters, each player chooses war or peace.
//	 Creatures controlled by players who chose war get +3/+0.
//	 Creatures controlled by players who chose peace get +0/+3."
//
// The choices are made in APNAP order as the Angel enters (CR 101.4) and are
// player notes on the permanent (#6097): kept while it stays, gone when it
// leaves (CR 400.7). The two statics read each creature's CURRENT controller
// against its note, so a creature that changes hands follows its new
// controller's choice, and a player with no note (one who joined later) gets
// neither bonus. Both are layer 7c modifications.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19e302f8-6ddb-4a43-98bc-ee71a456ac5a",
		Name:            "Archangel of Strife",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        effects.EachPlayerChoosesAsEnters("Archangel of Strife", "war", "peace"),
		Static: []game.StaticAbility{
			{
				Label:    "Creatures controlled by players who chose war get +3/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(c *game.Card, _ *game.Game, source *game.Card) bool {
					return c.IsCreature() && source.PlayerNoted(c.Controller, "war")
				},
				Apply: effects.PumpBy(3, 0),
			},
			{
				Label:    "Creatures controlled by players who chose peace get +0/+3.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(c *game.Card, _ *game.Game, source *game.Card) bool {
					return c.IsCreature() && source.PlayerNoted(c.Controller, "peace")
				},
				Apply: effects.PumpBy(0, 3),
			},
		},
	})
}
