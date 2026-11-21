package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endrek Sahr, Master Breeder — Legendary Creature — Human Wizard {4}{B},
// 2/2:
//
//	"Whenever you cast a creature spell, create X 1/1 black Thrull
//	 creature tokens, where X is that spell's mana value.
//	 When you control seven or more Thrulls, sacrifice Endrek Sahr."
//
// ADR 0107 §1 (#1858). The cast trigger reads the spell's mana value as
// it resolves (CR 608.2h), its announced X included while the spell is
// still on the stack. The sacrifice is a CR 603.8 state trigger over the
// Thrulls its controller controls — tokens or not, a changeling counts
// (CR 702.73a) — and it triggers once, however many Thrulls past seven
// arrive while it waits.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47a0079f-3544-45bc-a32a-bd93844c8c43",
		Name:         "Endrek Sahr, Master Breeder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Creature(), "Endrek Sahr — create X 1/1 black Thrulls",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.TriggeringSpellManaValue(g, item)
					if x <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 black Thrull"), N: x}.Apply(effects.NewContext(g, item))
				}),
			effects.WhenYouControlAtLeast(7, effects.QuerySubtype("Thrull"), "Endrek Sahr — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
