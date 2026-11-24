package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// King Harald's Revenge — Sorcery {2}{G}:
//
//	"Until end of turn, target creature gets +1/+1 for each creature you
//	 control and gains trample. It must be blocked this turn if able."
//
// The count is taken as the spell resolves (CR 608.2h) and fixed for the turn,
// the target itself counted if you control it. Irresistible Prey's block
// requirement is a data record pinned to the target until end of turn, and a
// target that has left (CR 608.2b) gets none of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc781549-7b2b-440b-93ae-69677acdfdad",
		Name:         "King Harald's Revenge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				n := effects.CountYouControl(ctx, effects.Creature())
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: n, Toughness: n,
					Label: "King Harald's Revenge — +1/+1 for each creature you control"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"},
					Label: "King Harald's Revenge — trample"}).Apply(ctx); err != nil {
					return err
				}
				return effects.BlockRequirementUntilEOT{Target: t.ID, Kind: game.BlockRequirementMustBeBlocked,
					Label: "King Harald's Revenge — must be blocked this turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
