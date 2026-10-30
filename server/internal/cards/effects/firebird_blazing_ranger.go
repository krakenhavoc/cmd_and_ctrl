package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firebird, Blazing Ranger — Legendary Creature — Human Ranger Hero
// {2}{R}, 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Firebird attacks, other attacking creatures get +X/+0 until
//	 end of turn, where X is Firebird's power."
//
// X is read as the trigger resolves (CR 608.2h: Firebird's last known
// power if it has left), and the set of attackers is locked then too
// (CR 611.2c), so a creature that joins the attack later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d917331-9f49-4d0c-bb2d-4853e3ff6ca1",
		Name:            "Firebird, Blazing Ranger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Firebird, Blazing Ranger — other attacking creatures get +X/+0, X is Firebird's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Match: And(AttackingCreature(), OtherThan(item.SourceCardID)),
						Power: sourcePowerNow(ctx),
						Label: "Firebird, Blazing Ranger",
					}.Apply(ctx)
				}),
		},
	})
}
