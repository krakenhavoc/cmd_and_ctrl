package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skygames — Enchantment — Aura {1}{U} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant land
//	 Enchanted land has "{T}: Target creature gains flying until end of turn.
//	 Activate only as a sorcery.""
//
// Debtor's Pulpit's granted row with a flying grant and the sorcery-speed gate
// (CR 602.5d): the land's controller, in their main phase with the stack empty.
//
// No simplification.
func init() {
	const grant = "skygames/flying"
	effects.Register(effects.Spec{
		OracleID:     "5eeb1495-4f0c-4920-8cd6-9c9257327e33",
		Name:         "Skygames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:        "{T}: Target creature gains flying until end of turn. Activate only as a sorcery.",
				Cost:         effects.TapCost(),
				Targets:      effects.TargetCreature("target creature"),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: "Skygames — flying until end of turn"}.Apply(ctx)
				},
			}},
			Text: "{T}: Target creature gains flying until end of turn. Activate only as a sorcery.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
