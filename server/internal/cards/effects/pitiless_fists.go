package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pitiless Fists — Enchantment — Aura {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Enchant creature you control
//	 When this Aura enters, enchanted creature fights up to one target
//	 creature an opponent controls.
//	 Enchanted creature gets +2/+2."
//
// The fight is the enchanted creature against the target, only while both
// are still on the battlefield (CR 701.14b); no target means no fight.
func init() {
	Register(Spec{
		OracleID:     "0d46a97a-b1d6-4bf3-b3ba-a5cdfd4c8bf5",
		Name:         "Pitiless Fists",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pitiless Fists — enchanted creature fights up to one target creature an opponent controls",
					pitilessFistsFight),
				TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1)),
		},
	})
}

func pitilessFistsFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return b10Fight(ctx, host, t.ID)
		}
	}
	return nil
}
