package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yeva's Forcemage — Creature — Elf Shaman {2}{G}, 2/2 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, target creature gets +2/+2 until end of
//	 turn."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// re-checked as it resolves (CR 608.2b); it may be the Forcemage itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "785e451d-87b6-4772-97e5-d2744818e0fd",
		Name:         "Yeva's Forcemage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Yeva's Forcemage — target creature gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2,
							Label: "Yeva's Forcemage — +2/+2"}.Apply(ctx)
					}
					return nil
				}), TargetCreature("target creature")),
		},
	})
}
