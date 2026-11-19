package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlwind Denial — Instant {2}{U} (Foundations Jumpstart, #5929):
//
//	"For each spell and ability your opponents control, counter it unless
//	 its controller pays {4}."
//
// Every spell, activated ability and triggered ability on the stack that an
// opponent controls gets its own CR 118.12 question for its controller,
// from the top of the stack down. Whirlwind Denial itself is yours, so it
// is never among them.
func init() {
	Register(Spec{
		OracleID:     "ba9e129d-a3cf-431c-93af-f2e4902d40a0",
		Name:         "Whirlwind Denial",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var objects []*game.StackItem
			for id, it := range ctx.Game.StackMeta {
				if it == nil || id == item.ID || it.Controller == ctx.Controller() {
					continue
				}
				switch it.Kind {
				case game.StackItemSpell, game.StackItemActivated, game.StackItemTriggered:
					objects = append(objects, it)
				}
			}
			sort.Slice(objects, func(i, j int) bool { return objects[i].Seq > objects[j].Seq })
			for _, it := range objects {
				if err := (CounterUnlessPaid{
					StackID:  it.ID,
					Cost:     "{4}",
					Question: "Whirlwind Denial — pay {4} or it is countered",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
