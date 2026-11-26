package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dual Casting — {1}{R} Aura: "Enchant creature. Enchanted creature has
// "{R}, {T}: Copy target instant or sorcery spell you control. You may
// choose new targets for the copy.""
//
// No simplification.
func init() {
	const grant = "dual-casting/copy"
	const text = "{R}, {T}: Copy target instant or sorcery spell you control. You may choose new targets for the copy."
	effects.Register(effects.Spec{
		OracleID:     "4d22b01b-63dd-4698-a221-0acaefb2000b",
		Name:         "Dual Casting",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetInstantOrSorcerySpell("target instant or sorcery spell you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CmdrB21CopyTheTargetSpell(item, effects.NewContext(g, item))
				},
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
