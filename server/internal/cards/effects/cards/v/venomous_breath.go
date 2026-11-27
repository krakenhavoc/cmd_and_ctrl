package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Venomous Breath — Instant {3}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"Choose target creature. At this turn's next end of combat, destroy all
//	 creatures that blocked or were blocked by it this turn."
//
// A delayed trigger (CR 603.7) for the beginning of the next end of combat step
// this turn. It is "next" strictly: the trigger remembers how many end of combat
// steps had already begun this turn and fires only at the first one after that,
// so a second combat does not destroy again. It expires at the end of the turn
// whether or not a combat follows. The creatures are read from the turn's block
// declarations when it fires: every creature that blocked the chosen creature and,
// if it blocked, every creature it blocked, whether or not it is still in
// combat, and each is destroyed if it is still on the battlefield. A chosen
// creature that has left changes nothing about who was blocked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc73bd1b-c713-4825-9f2d-f6ae95dddf09",
		Name:         "Venomous Breath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			return effects.ScheduleDelayedTrigger{
				At:           game.StepEndCombat,
				Label:        "Venomous Breath — destroy all creatures that blocked or were blocked by " + c.Name,
				EachThisTurn: true,
				Cards:        []uuid.UUID{id},
				Body:         venomousBreathBody,
				Params: game.EffectParams{
					Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch},
					Amount: effects.EndCombatsBegunThisTurn(ctx.Game),
				},
			}.Apply(ctx)
		},
	})
}

// venomousBreathBody destroys the creatures that blocked or were blocked by the
// chosen creature this turn, at the first end of combat after the spell resolved.
// An on-disk key: never renamed, never reused.
var venomousBreathBody = game.DelayedBody("venomous-breath/destroy-blockers-and-blocked",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if effects.EndCombatsBegunThisTurn(g) != p.Amount+1 {
			return nil
		}
		chosen := p.Object.ID
		seen := map[uuid.UUID]bool{}
		var victims []uuid.UUID
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind != game.EventBlock {
				continue
			}
			var other uuid.UUID
			switch chosen {
			case ev.Target:
				other = ev.CardID
			case ev.CardID:
				other = ev.Target
			default:
				continue
			}
			if other != uuid.Nil && !seen[other] {
				seen[other] = true
				victims = append(victims, other)
			}
		}
		ctx := effects.NewContext(g, item)
		for _, id := range victims {
			if !effects.OnBattlefield(g, id) {
				continue
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
