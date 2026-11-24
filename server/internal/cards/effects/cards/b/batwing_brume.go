package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Batwing Brume — Instant {1}{W/B} (Secret Lair Drop):
//
//	"Prevent all combat damage that would be dealt this turn if {W} was spent
//	 to cast this spell. Each player loses 1 life for each attacking creature
//	 they control if {B} was spent to cast this spell. (Do both if {W}{B} was
//	 spent.)"
//
// River's Grasp's mana-spent reading (ReadsManaSpent over the spell's own
// payment, so a waived payment asks what was paid, and a copy has no mana spent
// and does neither). The first half is Root Snare's prevention. The second counts
// the attacking creatures each player controls as it resolves (CR 608.2c, after
// the prevention, which does not touch life loss) and each loses that much.
// The hybrid symbol is paid with either colour, and WantsColors names both so an
// auto-paid cast reaches for both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fc8247f5-f7b2-4fc4-a9ea-90babad5fa42",
		Name:           "Batwing Brume",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W", "B"},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("W") > 0 {
				if err := (effects.PreventAllCombatDamageThisTurn{Label: "Batwing Brume: prevent combat damage"}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("B") == 0 {
				return nil
			}
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				n := len(effects.B13AttackingCreaturesYouControl(ctx.Game, p.ID))
				if n == 0 {
					continue
				}
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -n); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
