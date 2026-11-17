package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radha, Coalition Warlord — Legendary Creature — Elf Warrior {2}{R}{G},
// 3/3 (Dominaria United):
//
//	"Domain — Whenever Radha becomes tapped, another target creature you
//	 control gets +X/+X until end of turn, where X is the number of basic
//	 land types among lands you control."
//
// "Becomes tapped" is the tap event and the attack that taps it
// (ThisBecameTapped), so a vigilance attack is not one. X is the shared
// domain count (DomainCount, #5666) read as the ability resolves.
//
// No simplification.
func init() {
	const label = "Radha, Coalition Warlord — another target creature you control gets +X/+X until end of turn, where X is the number of basic land types among lands you control"
	pump := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		x := DomainCount(g, item.Controller)
		return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: label}.Apply(ctx)
	}
	Register(Spec{
		OracleID:     "fbf0bde7-cb13-402b-b3ab-81c787ce4827",
		Name:         "Radha, Coalition Warlord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(dmuB11BecameTapped(label, pump),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 0, 0))),
		},
	})
}
