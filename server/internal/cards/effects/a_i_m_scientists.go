package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A.I.M. Scientists — Creature — Human Scientist Villain {3}{U}, 3/3
// (Marvel Super Heroes, #4455):
//
//	"When this creature enters, it connives. (Draw a card, then discard
//	 a card. If you discarded a nonland card, put a +1/+1 counter on this
//	 creature.)
//	 Basic landcycling {2}"
//
// The enter trigger is connive.go's ConniveEach for the one creature that
// entered: the draw comes first, the discard is the controller's choice
// and waits for it, and the counter goes on only for a nonland card
// (CR 701.50a). If the Scientists have left by then the loot still
// happens and the counter is skipped (CR 701.50c). Basic landcycling is
// the cycling family's shared BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa732888-5b7d-4087-a6e5-1db1c98fbd4c",
		Name:         "A.I.M. Scientists",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("A.I.M. Scientists — it connives", Do(Instruction(func(ctx *Context) error {
				return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{ctx.Item.SourceCardID}}.Apply(ctx)
			}))),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
