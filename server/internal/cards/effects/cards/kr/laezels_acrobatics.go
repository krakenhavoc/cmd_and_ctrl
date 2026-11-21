package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lae'zel's Acrobatics — Instant {3}{W} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Exile all nontoken creatures you control, then roll a d20.
//	 1—9 | Return those cards to the battlefield under their owner's control at
//	 the beginning of the next end step.
//	 10—20 | Return those cards to the battlefield under their owner's control,
//	 then exile them again. Return those cards to the battlefield under their
//	 owner's control at the beginning of the next end step."
//
// One exile of every nontoken creature the caster controls (ExileAllMatching),
// whose continuation is handed the cards that reached exile and rolls the d20
// with them as data. 1—9 schedules the one return at the next end step (CR
// 603.7, the Ephemerate/Eerie Interlude return, as a new object). 10—20 returns
// them together now, so each sees the others enter (CR 603.6a) and any "enters"
// triggers happen, then the cards that came back (new objects, CR 400.7) are
// exiled again and scheduled to return at the next end step. A card no longer
// in exile when it should return is skipped (CR 603.7c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3efe6ac8-f92e-4108-913b-077e41c0a841",
		Name:         "Lae'zel's Acrobatics",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{
				Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Not(effects.IsTokenPredicate())),
				ThenTail: ctx.ItemTail(laezelsAcrobaticsExiled),
			}.Apply(ctx)
		},
	})
}

const laezelsAcrobaticsLabel = "Lae'zel's Acrobatics — return those cards to the battlefield at the beginning of the next end step"

// laezelsAcrobaticsExiled rolls the d20 over the cards that reached exile. An
// on-disk key.
var laezelsAcrobaticsExiled = effects.CardTail("then/laezels-acrobatics-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.RollDice{Sides: 20, N: 1, Then: laezelsAcrobaticsRolled, TailCarry: append([]uuid.UUID(nil), o.IDs...)}.Apply(ctx)
})

// laezelsAcrobaticsRolled reads the table. An on-disk key.
var laezelsAcrobaticsRolled = effects.CardTail("dice/laezels-acrobatics-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	if o.Rolls[0] < 10 {
		return effects.ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: laezelsAcrobaticsLabel,
			Cards: append([]uuid.UUID(nil), o.Tail.Carry...),
			Body:  effects.ReturnExiledToOwnersBody,
		}.Apply(ctx)
	}
	return effects.ReturnFromExileTogether{Targets: o.Tail.Carry, ThenTail: laezelsAcrobaticsReturned}.Apply(ctx)
})

// laezelsAcrobaticsReturned exiles the cards that came back, again, and
// schedules their return at the next end step. An on-disk key.
var laezelsAcrobaticsReturned = effects.CardTail("then/laezels-acrobatics-returned", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.ExileCardsThenScheduleReturn(ctx, o.IDs, laezelsAcrobaticsLabel)
})
