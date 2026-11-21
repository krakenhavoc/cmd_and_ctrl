package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Castle Doom — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast an
//	 artifact spell.
//	 {3}, {T}, Sacrifice an artifact: Create a 3/3 colorless Robot
//	 Villain artifact creature token named Doombot. Activate only as a
//	 sorcery."
//
// The second row is a restricted any-color mana ability (cast + Artifact
// tag, so it pays for artifact spells only and no ability). The third
// costs {3}, {T} and an artifact sacrificed as a cost (CR 118.3), at
// sorcery speed (CR 307.5). Castle Doom can sacrifice itself only if it
// were an artifact, which it is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bd013df-ad75-4099-940b-1765c58faf26",
		Name:         "Castle Doom",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "Add one mana of any color. Spend this mana only to cast an artifact spell.",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}, {T}, Sacrifice an artifact: Create a 3/3 colorless Robot Villain artifact creature token named Doombot. Activate only as a sorcery.",
			Cost: effects.Plus(effects.ManaCost("{3}"), game.AbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeSpec("an artifact", effects.Artifact()),
			}),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("3/3 colorless Robot Villain artifact named Doombot"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
