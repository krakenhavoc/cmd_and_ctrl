package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unravel — Instant {1}{U}{U} (Edge of Eternities, #4461):
//
//	"Counter target spell. If the amount of mana spent to cast that spell
//	 was less than its mana value, you draw a card."
//
// Mana Sculpt's counter with Tokka & Rahzar's comparison
// (tokkaAndRahzarSpentLessThanManaValue): the spell's own payment record
// against its mana value on the stack, X included (CR 202.3e). Both are
// read while the spell is still on the stack, before the counter, the
// last moment it is findable; neither can change once it was cast (CR
// 601.2h). A spell cast for less than its mana value (a cost reduction,
// convoke, an alternative cost, a free cast) draws the card; a tax or a
// kicker never lowers the amount below the mana value. A copy of a spell
// was not cast and spent nothing (CR 707.10), so countering one with a
// mana value above zero draws. The draw is not "if you do": a spell that
// can't be countered still draws.
//
// Like Mana Sculpt it declares ReadsAnotherSpellsAmount, which never asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "1525cbc1-0d95-4f58-99fb-136e54c4d486",
		Name:           "Unravel",
		ReadsManaSpent: game.ReadsAnotherSpellsAmount,
		Completeness:   CompletenessFull,
		Targets:        TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			spell := item.Targets[0].ID
			short := tokkaAndRahzarSpentLessThanManaValue(ctx.Game, spell)
			if err := (CounterTarget{StackID: spell}).Apply(ctx); err != nil {
				return err
			}
			if !short {
				return nil
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
