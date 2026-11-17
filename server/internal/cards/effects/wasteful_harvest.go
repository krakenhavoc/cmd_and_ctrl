package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wasteful Harvest — Instant {2}{G} (The Brothers' War):
//
//	"Mill five cards. You may put a permanent card from among the cards
//	 milled this way into your hand."
//
// Cache Grab's mill and pick (rfSpellBTakeFromMilledTail) at five cards:
// the pick is over the cards that actually reached the graveyard, up to
// one, so declining is an answer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d44627c0-1617-4706-bbeb-0169745364c7",
		Name:         "Wasteful Harvest",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), 5, game.ZoneGraveyard, game.MillUntil{},
				withWord(ctx.ItemTail(wastefulHarvestMilled), "Wasteful Harvest — put a permanent card milled this way into your hand"))
		},
	})
}

// wastefulHarvestMilled is the mill's keyed "then": the pick, nothing after
// it. An on-disk key: never renamed, never reused.
var wastefulHarvestMilled = rfSpellBTakeFromMilledTailPicked("mill-then-take/wasteful-harvest", func(_ *Context, _ []uuid.UUID) error { return nil })
