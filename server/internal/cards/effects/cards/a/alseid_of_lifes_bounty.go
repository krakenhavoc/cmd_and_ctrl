package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alseid of Life's Bounty — Enchantment Creature — Nymph {W}, 1/1 (Jumpstart 2022, #5948):
//
//	"Lifelink
//	 {1}, Sacrifice this creature: Target creature or enchantment you
//	 control gains protection from the color of your choice until end of
//	 turn."
//
// Mother of Runes's protection from a color of your choice, aimed at a
// creature or enchantment you control. The Alseid is sacrificed as a cost,
// so it can't be its own target as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d371031e-78be-4964-b521-b1eaa22a5480",
		Name:            "Alseid of Life's Bounty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Target creature or enchantment you control gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target creature or enchantment you control", effects.YouControl(), effects.Or(effects.Creature(), effects.Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				effects.ChooseProtectionColorFor(ctx, item.Controller, id, "Alseid of Life's Bounty", "Alseid of Life's Bounty — choose a color")
				return nil
			},
		}},
	})
}
