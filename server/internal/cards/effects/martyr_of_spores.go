package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyr of Spores — Creature — Human Shaman {G}, 1/1:
//
//	"{1}, Reveal X green cards from your hand, Sacrifice this creature:
//	 Target creature gets +X/+X until end of turn."
//
// The reveal cost is effects.RevealX (#2598, ADR 0020's 2026-10-08
// amendment): its count is the X announced with the activation, the
// revealed cards stay in the hand, and the mana cost is {1} whatever X
// is. The target is a creature, which may be any creature on the
// battlefield; if it left in response the ability does nothing, and the
// Martyr is gone either way, as printed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46c128d5-92b7-4132-a865-4db299879579",
		Name:         "Martyr of Spores",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Reveal X green cards from your hand, Sacrifice this creature: Target creature gets +X/+X until end of turn.",
			Cost:    Plus(ManaCost("{1}"), RevealX("X green cards", "G"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok || x <= 0 {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Martyr of Spores — +X/+X until end of turn"}.Apply(ctx)
			},
		}},
	})
}
