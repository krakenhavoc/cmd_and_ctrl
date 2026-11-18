package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bull's Strength — Instant {1}{G} (Adventures in the Forgotten Realms):
//
//	"Target creature gets +2/+2 and gains trample until end of turn. Untap it."
//
// One until-end-of-turn record for the pump and the keyword (CR 611.2a), then
// the untap. The untap is not part of the duration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5f10b54-10db-4c92-a596-8afea4e470f5",
		Name:         "Bull's Strength",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "trample")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := untilEndOfTurn(ctx, id, nil, "Bull's Strength — +2/+2 and trample",
				game.ModifyPTMod(2, 2), game.AddKeywordsMod("trample")); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
