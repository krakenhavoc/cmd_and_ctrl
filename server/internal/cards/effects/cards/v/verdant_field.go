package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Field — Enchantment — Aura {2}{G} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant land
//	 Enchanted land has '{T}: Target creature gets +1/+1 until end of turn.'"
//
// A granted activated ability on a land (Squirrel Nest's shape, ADR 0093)
// with Giant Growth's layer-7c until-end-of-turn boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d59fd2b-7fc6-49a7-8938-d3f08902a5a8",
		Name:         "Verdant Field",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: "verdant-field/pump",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gets +1/+1 until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.BoostUntilEOT{
						Target:    item.Targets[0].ID,
						Power:     1,
						Toughness: 1,
						Label:     "Verdant Field — +1/+1",
					}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Target creature gets +1/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("verdant-field/pump")},
	})
}
