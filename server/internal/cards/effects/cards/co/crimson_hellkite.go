package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Hellkite — Creature — Dragon {6}{R}{R}{R}, 6/6:
//
//	"Flying
//	 {X}, {T}: This creature deals X damage to target creature. Spend
//	 only red mana on X."
//
// The {X} is a cost-side spend restriction (#1600, SpendOnlyOnX("R")):
// X red symbols at payment, so a colourless or green X is refused by
// the pool, the auto-tapper and the bot's enumerator alike, and the
// largest X offered is the red mana the board can make with the
// Hellkite itself excluded (its own {T} is part of the cost). Under
// Chromatic Orrery any mana pays it (CR 609.4b).
//
// The {T} is the creature's own, so it is summoning-sick the turn it
// arrives (CR 302.6). The target is re-checked at resolution (CR
// 608.2b): a creature that left or gained hexproof in response takes
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37a1f340-10d5-407e-bf2b-cbbbb6cab985",
		Name:            "Crimson Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: This creature deals X damage to target creature. Spend only red mana on X.",
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.SpendOnlyOnX("R")),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
