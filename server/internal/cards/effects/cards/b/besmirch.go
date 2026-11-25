package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Besmirch — Sorcery {1}{R}{R} (Conspiracy: Take the Crown, S82 older sets
// batch 271, #5540):
//
//	"Until end of turn, gain control of target creature and it gains haste.
//	 Untap and goad that creature. (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you if
//	 able.)"
//
// Grab the Reins' theft (a layer 2 control change until end of turn) and
// haste grant, then the untap and your goad, in the printed order (CR
// 608.2c). The goad is the caster's and outlasts the theft: it ends as
// their next turn begins (CR 701.15a), so once the creature goes home it
// must attack each combat if able, and not the caster if another player can
// be attacked. A target that is illegal by resolution does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc80380c-7d75-4556-aaf2-bf7b2c757e75",
		Name:         "Besmirch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{
				Target:   id,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Besmirch — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Besmirch — haste"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.GoadAndScheduleClear(ctx, []uuid.UUID{id})
		},
	})
}
