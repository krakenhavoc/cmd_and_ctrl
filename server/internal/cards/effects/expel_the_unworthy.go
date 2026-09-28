package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expel the Unworthy — {1}{W} Sorcery:
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this
//	 spell.)
//	 Choose target creature with mana value 3 or less. If this spell was
//	 kicked, instead choose target creature. Exile the chosen creature,
//	 then its controller gains life equal to its mana value."
//
// The kicked clause replaces the printed one (#1716, WhenPaid). The
// controller and the mana value are read BEFORE the exile, as Swords
// to Plowshares reads its power: once the creature has left the
// battlefield it is a new object with no controller of its own.
func init() {
	Register(Spec{
		OracleID:      "4af2e62f-150e-4fd0-98b0-c6e72f5f9a51",
		Name:          "Expel the Unworthy",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{2}{W}"), TargetCreature("target creature"))},
		Targets:       TargetCreature("target creature with mana value 3 or less", ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			controller := card.Controller
			mv, _ := ctx.Game.ManaValueForEffect(card)
			if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: controller, Amount: mv}.Apply(ctx)
		},
	})
}
