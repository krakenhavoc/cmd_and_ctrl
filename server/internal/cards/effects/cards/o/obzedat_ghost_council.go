package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obzedat, Ghost Council — Legendary Creature — Spirit Advisor
// {1}{W}{W}{B}{B}, 5/5 (Gatecrash, S82 OLD b408, #5565):
//
//	"When Obzedat enters, target opponent loses 2 life and you gain 2 life.
//	 At the beginning of your end step, you may exile Obzedat. If you do, return
//	 it to the battlefield under its owner's control at the beginning of your
//	 next upkeep. It gains haste."
//
// The enters trigger names its opponent as it goes on the stack (CR 603.3d) and
// does both halves if the target is still legal (CR 608.2b). The end-step trigger
// is a "you may" (Optional), only for the same Obzedat that triggered it
// (CR 400.7), whose exile hands the cards that reached exile to a registered
// tail that schedules the return for the beginning of the controller's next
// upkeep (ControllerTurnOnly, CR 603.7). The return puts it onto the battlefield
// under its owner's control as a new object, which re-triggers the drain, and
// gives it haste for as long as it stays (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c3df703-e96a-4ee4-b449-b8e1e3460bd3",
		Name:         "Obzedat, Ghost Council",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Obzedat — target opponent loses 2 life and you gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := effects.TargetedPlayer(ctx)
					if victim == uuid.Nil {
						return nil
					}
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -2); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.Optional(effects.AtYourEndStep("Obzedat — you may exile it; return it at the beginning of your next upkeep with haste",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return g.ExileCardsKeyedForEffect([]uuid.UUID{item.SourceCardID}, ctx.ItemTail(obzedatScheduleReturn))
				}), "Obzedat — exile it until your next upkeep?"),
		},
	})
}

// obzedatScheduleReturn is "return it at the beginning of your next upkeep"
// for what reached exile. An on-disk key: never renamed, never reused.
var obzedatScheduleReturn = effects.CardTail("s82-old-b408/obzedat-schedule-return", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Obzedat — return it to the battlefield under its owner's control; it gains haste",
		Cards:              o.IDs,
		Body:               obzedatReturn,
	}.Apply(ctx)
})

// obzedatReturn returns the exiled cards under their owners' control; each
// that reached the battlefield is a new object, handed to obzedatHaste. An
// on-disk key: never renamed, never reused.
var obzedatReturn = game.SimpleDelayedBody("s82-old-b408/obzedat-return", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return effects.ReturnFromExileTogether{Targets: ids, ThenTail: obzedatHaste}.Apply(ctx)
})

// obzedatHaste is "it gains haste": each card that entered gets haste for as
// long as it stays on the battlefield (CR 611.2a). An on-disk key.
var obzedatHaste = effects.CardTail("s82-old-b408/obzedat-haste", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Obzedat — haste",
		}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
