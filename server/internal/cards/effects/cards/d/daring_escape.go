package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Escape — Instant {R} (Streets of New Capenna, #5491):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Scry 1."
//
// The scry is last: it queues a prompt the pump does not wait on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c777002d-dcd4-4d75-8175-5440a8bb118e",
		Name:         "Daring Escape",
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")), Scry: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				id := item.Targets[0].ID
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Daring Escape — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Daring Escape — first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
