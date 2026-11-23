package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bull's Strength — Instant {1}{G} (Adventures in the Forgotten Realms):
//
//	"Target creature gets +2/+2 and gains trample until end of turn. Untap it."
//
// One until-end-of-turn record for the pump and the keyword (CR 611.2a), then
// the untap. The untap is not part of the duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5f10b54-10db-4c92-a596-8afea4e470f5",
		Name:         "Bull's Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := effects.UntilEndOfTurn(ctx, id, nil, "Bull's Strength — +2/+2 and trample",
				game.ModifyPTMod(2, 2), game.AddKeywordsMod("trample")); err != nil {
				return err
			}
			return effects.UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
