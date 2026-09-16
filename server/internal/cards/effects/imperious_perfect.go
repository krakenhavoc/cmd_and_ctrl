package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperious Perfect — Creature — Elf Warrior {2}{G}, 2/2 (EDHREC rank
// 1430):
//
//	"Other Elves you control get +1/+1.
//	 {G}, {T}: Create a 1/1 green Elf Warrior creature token."
//
// The Elf lord that makes its own Elves. The anthem is a Layer 7c
// static over OTHER Elves the controller controls — effective
// subtypes, so a changeling counts and the Perfect itself does not,
// as printed — and each token it makes is an Elf, so it is pumped
// the moment it lands. The tap ability is a creature's {T}, so
// summoning sickness applies (CR 302.6); the engine enforces it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fa71348-fa4d-4f39-a451-cf1570591991",
		Name:         "Imperious Perfect",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype("Elf")
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{G}, {T}: Create a 1/1 green Elf Warrior creature token.",
			Cost:  Plus(ManaCost("{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Elf Warrior"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
