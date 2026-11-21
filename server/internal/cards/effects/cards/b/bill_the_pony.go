package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bill the Pony — Legendary Creature — Horse {3}{W}, 1/4 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"When Bill the Pony enters, create two Food tokens.
//	 Sacrifice a Food: Until end of turn, target creature you control
//	 assigns combat damage equal to its toughness rather than its power."
//
// Bag End Banquet's Food tokens and Bosco's "Sacrifice a Food" cost (any
// Food artifact you control, token or not). The damage rule is the
// until-end-of-turn mod Walking Bulwark grants (DamageByToughnessMod), on
// a target creature still on the battlefield as the ability resolves (CR
// 608.2b); it applies whatever the creature's power is (CR 510.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e461ef8d-59ca-440a-8773-da1de3b6b195",
		Name:         "Bill the Pony",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bill the Pony — create two Food tokens", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Food: Until end of turn, target creature you control assigns combat damage equal to its toughness rather than its power.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))},
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, id, nil, "Bill the Pony — assigns combat damage equal to its toughness", game.DamageByToughnessMod())
			},
		}},
	})
}
