package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "47a0079f-3544-45bc-a32a-bd93844c8c43",
		Name:         "Endrek Sahr, Master Breeder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Endrek Sahr — create X 1/1 black Thrulls",
				func(g *game.Game, item *game.StackItem) error {
					x := triggeringSpellManaValue(g, item)
					if x <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 black Thrull"), N: x}.Apply(NewContext(g, item))
				}),
			WhenYouControlAtLeast(7, QuerySubtype("Thrull"), "Endrek Sahr — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
