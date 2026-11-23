package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spinal Embrace — Instant {3}{U}{U}{B} (Dominaria Remastered, #5948):
//
//	"Cast this spell only during combat.
//	 Untap target creature you don't control and gain control of it. It
//	 gains haste until end of turn. At the beginning of the next end step,
//	 sacrifice it. If you do, you gain life equal to its toughness."
//
// The control change has no duration (CR 611.2). The sacrifice is a delayed
// trigger that names the creature; it sacrifices it only if you still
// control it, and the life is its toughness as it left the battlefield
// (last-known information, CR 608.2h), gained only if it was sacrificed.
//
// No simplification.
var spinalEmbraceSacrifice = game.SimpleDelayedBody("s82-gap-b141/spinal-embrace-sacrifice", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !effects.OnBattlefield(g, t.ID) || c.Controller != item.Controller {
			continue
		}
		if err := (effects.SacrificePermanent{Target: t.ID, ThenTail: ctx.TailN(spinalEmbraceLife, c.CurrentToughness())}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

var spinalEmbraceLife = effects.CardTail("s82-gap-b141/spinal-embrace-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cf3fb65-9107-428a-8853-029ec97112b5",
		Name:         "Spinal Embrace",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat
		},
		CastConditionLabel: "Cast this spell only during combat.",
		Targets:            effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GainControl{
				Target:     target,
				Controller: item.Controller,
				Duration:   game.IndefiniteDuration(),
				Label:      "Spinal Embrace — gain control",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Spinal Embrace — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: item.Controller,
				Label:      "Spinal Embrace — sacrifice it, then gain life equal to its toughness",
				Cards:      []uuid.UUID{target},
				Body:       spinalEmbraceSacrifice,
			}.Apply(ctx)
		},
	})
}
