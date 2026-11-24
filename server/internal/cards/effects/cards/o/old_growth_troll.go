package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old-Growth Troll — Creature — Troll Warrior {G}{G}{G}, 4/4:
//
//	"Trample
//	 When Old-Growth Troll dies, if it was a creature, return it to the
//	 battlefield. It's an Aura enchantment with enchant Forest you control and
//	 'Enchanted Forest has "{T}: Add {G}{G}" and "{1}, {T}, Sacrifice this land:
//	 Create a tapped 4/4 green Troll Warrior creature token with trample."'"
//
// Harold and Bob's returned-Aura seam (returned_aura.go): the dies trigger is
// WhenThisDiesReturnItAsAnAura, read from the last-known information, the Aura
// form is ReturnedAuraForm over the grant bundle below, and the Troll is
// held to Forest you control (CR 303.4f, 704.5m). The bundle gives the Forest a
// mana ability producing two green mana and an activated ability that
// sacrifices the Forest and makes the tapped 4/4 Troll Warrior token with trample.
//
// No simplification.
func init() {
	const grant = "old-growth-troll/two-green-and-sacrifice-for-troll"
	effects.Register(effects.Spec{
		OracleID:        "11e3f1d3-9d78-4c00-b965-e14407fd5210",
		Name:            "Old-Growth Troll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ReturnedAura:    effects.EnchantLand(effects.Subtype("Forest"), effects.YouControl()),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add {G}{G}\n{1}, {T}, Sacrifice this land: Create a tapped 4/4 green Troll Warrior creature token with trample.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}{G}",
				Label:    "Add {G}{G}",
			}},
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, {T}, Sacrifice this land: Create a tapped 4/4 green Troll Warrior creature token with trample.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAdvanced{
						Controller: item.Controller,
						Spec:       effects.Token(effects.TokenCard("4/4 green Troll Warrior with trample")).EntersTapped(),
						N:          1,
					}.Apply(effects.NewContext(g, item))
				},
			}},
		}},
		Static: effects.ReturnedAuraForm(grant),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDiesReturnItAsAnAura("Old-Growth Troll"),
		},
	})
}
