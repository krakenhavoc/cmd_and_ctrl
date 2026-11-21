package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memory Jar — Artifact {5} (Urza's Legacy, #5841):
//
//	"{T}, Sacrifice this artifact: Each player exiles all cards from their
//	 hand face down and draws seven cards. At the beginning of the next end
//	 step, each player discards their hand and returns to their hand each
//	 card they exiled this way."
//
// Tapping and sacrificing the Jar are the cost, paid as the ability is
// activated (CR 602.2b, 601.2h), so it is in the graveyard before the
// ability resolves and nothing it does can reach it.
//
// Each player's hand is exiled face down in APNAP order, one hand at a time
// (eachPlayerInTurn), so a CR 616 prompt on one hand holds the next behind
// it and is a restore point. The cards land FaceDownExiled
// (Game.ExileHandFaceDownKeyedForEffect): CR 406.3 lets no player examine a
// card exiled face down unless an instruction allows it, and this one does
// not, so nobody can look at them, their owner included. Then every player
// draws seven, after any prompt the exile asked (CR 608.2c).
//
// Once the last hand is exiled, the chain's tail schedules the CR 603.7
// delayed trigger with every card exiled this way and the CR 400.7 epoch
// each landed with, as plain data (EffectParams.Objects). "The next end
// step" is any player's: activated in an end step, it waits for the next
// turn's. When it resolves every player discards their whole hand (a real
// discard, so madness and discard triggers see it, and the Jar is the
// discard's source), and then each card is returned to its owner's hand
// only if it is still the object it was when it was exiled: a card that
// left exile in between, even one that came back to exile, is a new object
// and stays where it is (CR 400.7, CR 603.7c). The return waits for any
// prompt a discard asked. A card returns to the hand of its owner, who is
// the player who exiled it (CR 400.3). The trigger is scheduled even when
// every hand was empty: the discard still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     MemoryJarOracle,
		Name:         "Memory Jar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice this artifact: Each player exiles all cards from their hand face down and draws seven cards. At the beginning of the next end step, each player discards their hand and returns to their hand each card they exiled this way.",
			// A whole-table hand swap answers nothing on the stack (ADR
			// 0142). No Draws amount: the seven go back at the end step.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.EachPlayerInTurn(ctx.Game, ctx.ItemTail(memoryJarExileStep), effects.ApnapPlayers(ctx.Game), nil,
						ctx.ItemTail(memoryJarScheduleReturn), memoryJarExileAct)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					for _, id := range effects.ApnapPlayers(ctx.Game) {
						if err := (effects.DrawCards{Player: id, N: 7}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			),
		}},
	})
}

const MemoryJarOracle = "eb9c9a34-88a8-4afe-a053-806f6b29834f"

// memoryJarExileStep is one step of the each-player exile chain. An on-disk
// key: never renamed, never reused.
var memoryJarExileStep = effects.RegisterEachPlayerStep("memory-jar/exile-hand-step", memoryJarExileAct)

// memoryJarExileAct exiles one player's whole hand face down.
func memoryJarExileAct(ctx *effects.Context, player uuid.UUID, _ int, step game.KeyedTail) error {
	return ctx.Game.ExileHandFaceDownKeyedForEffect(player, step)
}

// memoryJarScheduleReturn runs once every hand is exiled, told every card
// exiled this way: it schedules the end-step discard and return over
// exactly those objects. An on-disk key: never renamed, never reused.
var memoryJarScheduleReturn = effects.CardTail("memory-jar/schedule-return", func(ctx *effects.Context, o game.TailOutcome) error {
	var objs []game.ObjectRef
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			objs = append(objs, game.ObjectRef{ID: id, Epoch: ctx.Game.ObjectEpochForEffect(id)})
		}
	}
	return effects.ScheduleDelayedTrigger{
		At:     game.StepEnd,
		Label:  "Memory Jar — each player discards their hand and returns to their hand each card they exiled this way",
		Body:   memoryJarReturnBody,
		Params: game.EffectParams{Objects: objs},
	}.Apply(ctx)
})

// memoryJarReturnBody is the delayed trigger: discard every hand, then
// return the cards still exiled. An on-disk key: never renamed, never
// reused.
var memoryJarReturnBody = game.DelayedBody("memory-jar/discard-and-return", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	return memoryJarDiscardThenReturn.Run(effects.NewContext(g, item))
})

// memoryJarDiscardThenReturn is the delayed trigger's two instructions in
// the order written (CR 608.2c): the return waits for a prompt a discard
// asked (madness, a CR 616 order). Each step reads the trigger's own item,
// so a table restored between them still returns the cards. An on-disk key.
var memoryJarDiscardThenReturn = effects.RegisterThenSteps("then-steps/memory-jar-discard-and-return",
	effects.Instruction(func(ctx *effects.Context) error {
		for _, id := range effects.ApnapPlayers(ctx.Game) {
			p := ctx.Game.PlayerByIDForEffect(id)
			if p == nil || p.Hand.Size() == 0 {
				continue
			}
			hand := make([]uuid.UUID, len(p.Hand.Cards))
			for i, c := range p.Hand.Cards {
				hand[i] = c.InstanceID
			}
			if err := ctx.Game.DiscardCardsKeyedForEffect(id, hand, ctx.Source(), game.KeyedTail{}); err != nil {
				return err
			}
		}
		return nil
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		var back []uuid.UUID
		for _, ref := range ctx.Item.Params.Objects {
			z := ctx.Game.FindCardZoneForEffect(ref.ID)
			if z != nil && z.Kind == game.ZoneExile && ctx.Game.ObjectEpochForEffect(ref.ID) == ref.Epoch {
				back = append(back, ref.ID)
			}
		}
		if len(back) > 0 {
			ctx.Game.BounceCardsToHandForEffect(back)
		}
		return nil
	}),
)
