package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aragorn, King of Gondor — Legendary Creature — Human Noble {1}{U}{R}{W},
// 4/4 (Tales of Middle-earth Commander):
//
//	"Vigilance, lifelink
//	 When Aragorn enters, you become the monarch.
//	 Whenever Aragorn attacks, up to one target creature can't block this
//	 turn. If you're the monarch, creatures can't block this turn."
//
// The attack trigger is one effect: the (optional) target first, then the
// monarch clause read as it resolves (not an intervening "if"). "Creatures
// can't block" is a rule restriction on every creature, read live until
// end of turn, so creatures that enter later cannot block either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be97c400-c690-426d-a39e-4082080a3823",
		Name:            "Aragorn, King of Gondor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aragorn, King of Gondor — you become the monarch", Do(BecomeTheMonarch{})),
			Targeting(WheneverThisAttacks("Aragorn, King of Gondor — up to one target creature can't block this turn; if you're the monarch, creatures can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Aragorn, King of Gondor — can't block"}).Apply(ctx); err != nil {
							return err
						}
					}
					if !YoureTheMonarch(g, item.Controller) {
						return nil
					}
					return RestrictUntilEOT{Scope: game.ScopeCreatures, Restrictions: game.CantBlock, Label: "Aragorn, King of Gondor — creatures can't block"}.Apply(ctx)
				}),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}
