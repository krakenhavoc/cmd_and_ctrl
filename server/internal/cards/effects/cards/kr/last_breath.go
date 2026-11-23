package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last Breath — Instant {1}{W} (S82 GAP, #5948):
//
//	"Exile target creature with power 2 or less. Its controller gains 4 life."
//
// Swords to Plowshares' shape: the controller is read before the exile, as the
// creature's last-known controller (CR 608.2h). The power restriction is part of
// the target spec, re-checked as the spell resolves (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6fb4f0d-ce49-4913-a2ef-88365e4916cf",
		Name:         "Last Breath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := card.Controller
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, 4)
		},
	})
}
