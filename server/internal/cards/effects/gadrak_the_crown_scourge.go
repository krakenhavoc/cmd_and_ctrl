package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gadrak, the Crown-Scourge — Legendary Creature — Dragon {2}{R}, 5/4
// (Tarkir: Dragonstorm Commander):
//
//	"Flying
//	 Gadrak can't attack unless you control four or more artifacts.
//	 At the beginning of your end step, create a Treasure token for each
//	 nontoken creature that died this turn."
//
// The attack restriction is Chief Warg's Company's shape: a layer-6 static
// on the creature itself that sets CantAttack while its controller controls
// fewer than four artifacts, read live after the layers. It only gates the
// declaration, so a Gadrak that has attacked stays attacking when the
// artifacts go (the printed ruling). The end step trigger reads the
// table-wide tally of nontoken creatures that died this turn
// (TurnTally.NontokenCreaturesDied), which counts those that died before
// Gadrak entered, whoever controlled them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "42d7ec93-0943-43e2-9b65-b7edc3349410",
		Name:            "Gadrak, the Crown-Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label: "Gadrak can't attack unless you control four or more artifacts.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) &&
					len(permanentsControlledByMatching(g, source.Controller, Artifact())) < 4
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Gadrak, the Crown-Scourge — create a Treasure token for each nontoken creature that died this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.TurnTally.NontokenCreaturesDied
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TreasureToken(), N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
