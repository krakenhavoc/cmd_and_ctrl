package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torment of Hailfire — Sorcery {X}{B}{B}:
//
//	"Repeat the following process X times. Each opponent loses 3 life
//	 unless that player sacrifices a nonland permanent of their choice
//	 or discards a card."
//
// The card that forced the option pick (#568). Each repetition puts a
// THREE-WAY question in front of every opponent, and none of the three
// prompts that existed could ask it: pay_unless is a mana payment,
// confirm is two branches, and a trigger's "you may" belongs to the
// trigger's controller.
//
// The run of questions — built from what each player can do, and asked
// one at a time, each queued by the answer to the one before — is the
// shared punisherRepeat (punisher_repeat.go), which Rottenmouth Viper
// asks once per blight counter.
//
// X is the announced X (CR 601.2b), read off the stack item.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30a0a6c2-1fbb-4784-ab96-22611d57e62c",
		Name:         "Torment of Hailfire",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return tormentOfHailfire.start(ctx, ctx.X())
		},
	})
}

// tormentOfHailfire is the printed punisher: 3 life a refusal.
var tormentOfHailfire = punisherRepeat{Name: "Torment of Hailfire", Life: 3}
