package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Racecourse Fury — Enchantment — Aura {R} (Return to Ravnica, S82 OLD):
//
//	"Enchant land
//	 Enchanted land has "{T}: Target creature gains haste until end of
//	 turn.""
//
// A granted activated ability on a land (Squirrel Nest's shape), activated by
// the land's controller, who is not always the Aura's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7093835-0f84-4ea4-b228-2e39bd2642ab",
		Name:         "Racecourse Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key: racecourseFuryGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gains haste until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Racecourse Fury — haste until end of turn"}.Apply(ctx)
				},
			}},
			Text: "{T}: Target creature gains haste until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(racecourseFuryGrant)},
	})
}

// racecourseFuryGrant is the granted ability's bundle key: an on-disk
// identity, never renamed, never reused.
const racecourseFuryGrant = "racecourse-fury/grant-haste"
