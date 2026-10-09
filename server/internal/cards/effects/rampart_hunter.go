package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampart Hunter — Creature — Horror {3}{B}, 3/3:
//
//	"Deathtouch
//	 When this creature enters, target creature gets +2/+2 and gains
//	 deathtouch until end of turn."
//
// The target is any creature, this one included, chosen as the trigger
// goes on the stack. Both halves lock to that creature at resolution.
//
// No simplification.
func init() {
	const label = "Rampart Hunter — target creature gets +2/+2 and gains deathtouch until end of turn"
	Register(Spec{
		OracleID:        "0867755a-9fb9-42ad-897e-6206fbfbc1f8",
		Name:            "Rampart Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				if err := (BoostUntilEOT{Target: ts[0].ID, Power: 2, Toughness: 2, Label: "Rampart Hunter — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"deathtouch"}, Label: "Rampart Hunter — deathtouch"}.Apply(ctx)
			}), TargetCreature("target creature")),
		},
	})
}
