package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Settle the Score — Sorcery {2}{B}{B} (Jumpstart, #5510):
//
//	"Exile target creature. Put two loyalty counters on a planeswalker you
//	 control."
//
// The planeswalker is not targeted: it is chosen as the spell resolves,
// among those you control then (CR 608.2d), and the choice is skipped when
// there is only one. The counters go on whatever the player picks even if
// the creature was gone, as long as the spell still resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f18bb2d-249f-4d7a-a0f7-8a285d365777",
		Name:         "Settle the Score",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileFirstLegalCardTarget(ctx.Game, item); err != nil {
				return err
			}
			g := ctx.Game
			var walkers []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsPlaneswalker() {
					walkers = append(walkers, c.InstanceID)
				}
			}
			switch len(walkers) {
			case 0:
				return nil
			case 1:
				return effects.NothingIfGone(g.AddCounterForEffect(walkers[0], game.CounterLoyalty, 2))
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Settle the Score — put two loyalty counters on a planeswalker you control",
				Cards:    walkers,
				Min:      1,
				Max:      1,
				ThenKey:  settleTheScorePick,
				LiveZone: game.ZoneBattlefield,
			})
			return nil
		},
	})
}

// settleTheScorePick puts the two loyalty counters on the chosen planeswalker.
// An on-disk key: never renamed, never reused.
var settleTheScorePick = game.RegisterChooseCardsThen("choose-cards/s82-old-b103-settle-the-score",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		return effects.NothingIfGone(g.AddCounterForEffect(r.Picked[0], game.CounterLoyalty, 2))
	})
