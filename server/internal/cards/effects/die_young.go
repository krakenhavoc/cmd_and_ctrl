package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Die Young — Sorcery {1}{B}:
//
//	"Choose target creature. You get {E}{E} (two energy counters), then
//	 you may pay any amount of {E}. The creature gets -1/-1 until end of
//	 turn for each {E} paid this way."
//
// ADR 0129 §3 (#1995, owner decision 3): the pay_amount prompt, and one
// -X/-X until end of turn for the X paid. The stepper and the bot start
// at the amount that kills the creature: its toughness less its marked
// damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bcc54d1-8ab6-4ad3-91b6-97fb02b88cc0",
		Name:         "Die Young",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyAmount{
				Question: "Die Young — pay any amount of {E}; the creature gets -1/-1 for each",
				Unit:     game.PayAmountPower,
				Goal:     lethalDamageGoal,
				Then:     dieYoungShrinks,
			}.Apply(ctx)
		},
	})
}

// dieYoungShrinks is the paid half (#4294): -1/-1 until end of turn for
// each {E} paid, on the first target of the item that asked.
var dieYoungShrinks = NumberThen("number/die-young-minus-one-for-each",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 || ctx.Item == nil || len(ctx.Item.Targets) == 0 {
			return nil
		}
		return BoostUntilEOT{Target: ctx.Item.Targets[0].ID, Power: -r.Number, Toughness: -r.Number,
			Label: "Die Young — -1/-1 for each {E} paid"}.Apply(ctx)
	})
