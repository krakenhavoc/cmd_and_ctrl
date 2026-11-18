package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodsworn Steward — Creature — Vampire Knight {2}{R}{R}, 4/4:
//
//	"Flying
//	 Commander creatures you control get +2/+2 and have haste."
//
// Bastion Protector's statics, one per layer: +2/+2 in layer 7c and haste
// in layer 6, over the same "commander creature you control" predicate. It
// pumps itself when it is the commander.
//
// No simplification.
func init() {
	commanderCreatureYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() &&
			target.IsCommander &&
			target.Controller == source.Controller
	}
	Register(Spec{
		OracleID:        "a5ec7ccd-d1fe-4ca6-b019-d4616c233771",
		Name:            "Bloodsworn Steward",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			{
				Label:     "Commander creatures you control get +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: commanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:     "Commander creatures you control have haste.",
				Layer:     game.Layer6Ability,
				AppliesTo: commanderCreatureYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = append(c.Abilities, "haste")
				},
			},
		},
	})
}
