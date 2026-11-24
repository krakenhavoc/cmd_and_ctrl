package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rain of Filth — Instant {B}:
//
//	"Until end of turn, lands you control gain 'Sacrifice this land: Add {B}.'"
//
// A resolving spell's ability grant (GrantAbilitiesFor, Feign Death's seam): the
// granted mana ability is a bundle this card registers, its cost the sacrifice
// of the land it is on. The lands are the ones you control as the spell
// resolves (CR 611.2c); a land that enters later this turn doesn't get it, and
// one that leaves and returns is a new object without it (CR 400.7). It is a
// mana ability (CR 605.1a), so it doesn't use the stack and can be used in the
// middle of paying a cost; the land is sacrificed as the cost.
//
// No simplification.
func init() {
	const key = "rain-of-filth/sacrifice-for-black"
	effects.Register(effects.Spec{
		OracleID:     "aa77980d-0434-40f0-990f-1c8c50610174",
		Name:         "Rain of Filth",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  key,
			Text: "Sacrifice this land: Add {B}.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Sacrifice: true},
				Produced: "{B}",
				Label:    "Sacrifice this land: Add {B}.",
			}},
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.And(effects.Land(), effects.YouControl()),
				Keys:  []string{key},
				Label: "Rain of Filth — lands you control gain \"Sacrifice this land: Add {B}.\"",
			}.Apply(ctx)
		},
	})
}
