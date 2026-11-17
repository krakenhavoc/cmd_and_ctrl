package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refuse to Yield — Instant {1}{W} (Streets of New Capenna, #5491):
//
//	"Target creature gets +2/+7 until end of turn. Untap it."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d05aeada-2a9f-464d-8a66-b6e90f43bae3",
		Name:         "Refuse to Yield",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 7)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 7, Label: "Refuse to Yield — +2/+7"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
