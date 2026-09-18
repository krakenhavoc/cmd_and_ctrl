package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Field Marshal — Creature — Angel {2}{W}{W}, 3/3 (EDHREC
// rank 4485):
//
//	"Flying
//	 Lieutenant — As long as you control your commander, this creature
//	 gets +2/+2 and creatures you control have vigilance."
//
// The white lieutenant from Commander 2014, and a cleaner one than
// most: a 5/5 flier for four in any deck whose commander is on the
// battlefield, plus a team-wide vigilance that makes an alpha strike
// safe. It rewards exactly the thing Commander already wants you to
// do.
//
// "You control your commander" is b18ControlsYourCommander — a
// commander the Marshal's controller OWNS and controls. A stolen
// opposing commander does not switch the lieutenant on (it is not
// YOUR commander), and your own commander stolen by an opponent
// switches it off. The check re-runs on every layer recompute, and a
// commander's arrival or departure is a battlefield zone move, which
// is what invalidates the cache.
//
// Two statics, because the two halves live in different layers and
// one entry could not sort into both: a layer 7c +2/+2 pinned to the
// Marshal itself, and a layer 6 vigilance grant over every creature
// its controller has — the Marshal INCLUDED, because "creatures you
// control" has no "other".
//
// No simplification.
func init() {
	lieutenant := func(g *game.Game, source *game.Card) bool {
		return b18ControlsYourCommander(g, source.Controller)
	}
	Register(Spec{
		OracleID:        "502575f4-7c56-44bc-a77f-ae28d66c8e1f",
		Name:            "Angelic Field Marshal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && lieutenant(g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && lieutenant(g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !keywordSliceContains(c.Abilities, "vigilance") {
						c.Abilities = append(c.Abilities, "vigilance")
					}
				},
			},
		},
	})
}
