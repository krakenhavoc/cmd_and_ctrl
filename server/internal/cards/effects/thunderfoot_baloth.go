package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderfoot Baloth — Creature — Beast {4}{G}{G}, 5/5 (EDHREC rank
// 2010):
//
//	"Trample
//	 Lieutenant — As long as you control your commander, this
//	 creature gets +2/+2 and other creatures you control get +2/+2
//	 and have trample."
//
// The Commander 2014 lieutenant. "You control your commander" is
// read at every recompute — a commander the controller OWNS and
// controls (b18ControlsYourCommander; a stolen opposing commander
// does not count) — and the commander's arrival or departure is a
// battlefield zone move, which invalidates the layer cache. Two
// statics: a layer 7c anthem over every creature you control (the
// Baloth included — the two +2/+2 clauses add up to one), and a
// layer 6 trample grant over the OTHERS, since the Baloth has its
// own.
//
// No simplification.
func init() {
	lieutenant := func(_ *game.Card, g *game.Game, source *game.Card) bool {
		return b18ControlsYourCommander(g, source.Controller)
	}
	Register(Spec{
		OracleID:        "f55334d9-1ec2-4667-bf28-5a0384d6f053",
		Name:            "Thunderfoot Baloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && lieutenant(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller &&
						target.InstanceID != source.InstanceID && lieutenant(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, k := range c.Abilities {
						if k == "trample" {
							return
						}
					}
					c.Abilities = append(c.Abilities, "trample")
				},
			},
		},
	})
}
