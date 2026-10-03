package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "37a1f340-10d5-407e-bf2b-cbbbb6cab985",
		Name:            "Crimson Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: This creature deals X damage to target creature. Spend only red mana on X.",
			Cost:    Plus(ManaCost("{X}"), TapCost(), SpendOnlyOnX("R")),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: id, Amount: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
