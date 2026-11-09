package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thickest in the Thicket — Enchantment {3}{G}{G}:
//
//	"When this enchantment enters, put X +1/+1 counters on target creature,
//	 where X is that creature's power.
//	 At the beginning of your end step, draw two cards if you control the
//	 creature with the greatest power or tied for the greatest power."
//
// X is the target's power as the ability resolves, counted once (CR 608.2b;
// a target that has left, or whose power is not positive, gets nothing).
// The end-step ability has no intervening "if": it triggers every end step
// and the condition is judged as it resolves (the ruling), over every
// creature on the battlefield, whoever controls it. Controlling no
// creature at all is not controlling the greatest power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d60dae0-7be4-4859-99ac-781a27bcfd93",
		Name:         "Thickest in the Thicket",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Thickest in the Thicket — put X +1/+1 counters on target creature, where X is that creature's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						c, ok := g.LookupCardForEffect(t.ID)
						if !ok || !onBattlefield(g, t.ID) {
							return nil
						}
						x := c.CurrentPower()
						if x <= 0 {
							return nil
						}
						return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
					}
					return nil
				}), TargetCreature("target creature")),
			AtYourEndStep("Thickest in the Thicket — draw two cards if you control the creature with the greatest power or tied for the greatest power",
				Do(If{
					Cond: CondBoard{Name: "you control the creature with the greatest power or tied for it", Test: func(ctx *Context) bool {
						return controlsGreatestPower(ctx.Game, ctx.Controller())
					}},
					Then: []Applier{DrawCards{N: 2}},
				})),
		},
	})
}

// controlsGreatestPower is "you control the creature with the greatest
// power or tied for the greatest power": some creature of player's has a
// power no creature at the table beats. Powers are read through the layers
// (CR 208.1, negative power included).
func controlsGreatestPower(g *game.Game, player uuid.UUID) bool {
	best := 0
	have := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			if p := c.PowerForComparison(); !have || p > best {
				best, have = p, true
			}
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == player && c.PowerForComparison() == best {
			return true
		}
	}
	return false
}
