package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief Warg's Company — Creature — Wolf {1}{B}{G}, 5/3 (The Hobbit, HOB):
//
//	"Trample
//	 This creature can't attack unless you control two or more other Wolves.
//	 At the beginning of your upkeep, create a 2/2 green Wolf creature
//	 token."
//
// The attack restriction is Hazoret's shape: a layer 6 static on the
// creature itself that sets CantAttack while its controller controls fewer
// than two OTHER Wolves (the Company is a Wolf itself and does not count),
// read live after the layers, so a changeling counts and the restriction
// lifts the moment the second Wolf arrives. The upkeep trigger makes a Wolf
// each of its controller's turns, which is what pays for the restriction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "372d7086-b7dc-4b6d-bdf6-e9ed9c995456",
		Name:            "Chief Warg's Company",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label: "Can't attack unless you control two or more other Wolves.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && hobA9OtherSubtypeYouControl(g, source, "Wolf") < 2
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Chief Warg's Company — create a 2/2 green Wolf creature token",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1})),
		},
	})
}
