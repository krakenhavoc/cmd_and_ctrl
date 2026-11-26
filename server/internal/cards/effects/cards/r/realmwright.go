package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Realmwright — Creature — Vedalken Wizard {U}, 1/1 (Gatecrash, #5565):
//
//	"As this creature enters, choose a basic land type.
//	 Lands you control are the chosen type in addition to their other types."
//
// The choice is made as it enters (CR 614.12, Convincing Mirage's
// ChooseBasicLandTypeAsEnters) and stored on the permanent; five statics, one
// per basic land type, are gated on that type being the chosen one
// (StaticWhenChosen), so exactly one applies once the choice is made. Each is
// a layer 4 addition to the subtypes of every land its controller controls,
// kept beside the lands' other types (CR 305.7 is the "becomes" form; this
// one adds, so abilities stay), and the added type brings its intrinsic mana
// ability (CR 305.6). It stops when the Realmwright leaves.
//
// No simplification.
func init() {
	var statics []game.StaticAbility
	for _, t := range game.BasicLandTypes {
		land := t
		statics = append(statics, effects.StaticWhenChosen(land, game.StaticAbility{
			Label: "Lands you control are " + land + " in addition to their other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, st := range c.Subtypes {
					if st == land {
						return
					}
				}
				c.Subtypes = append(c.Subtypes, land)
			},
		}))
	}
	effects.Register(effects.Spec{
		OracleID:     "bad869dd-e844-46ca-be53-73c4bc19d09c",
		Name:         "Realmwright",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseBasicLandTypeAsEnters("Realmwright"),
		Static:       statics,
	})
}
