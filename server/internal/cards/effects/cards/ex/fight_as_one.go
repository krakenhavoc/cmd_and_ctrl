package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fight as One — Instant {W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Choose one or both —
//	 • Target Human creature you control gets +1/+1 and gains indestructible
//	   until end of turn.
//	 • Target non-Human creature you control gets +1/+1 and gains
//	   indestructible until end of turn."
//
// Each bullet has its own target group, read by its occurrence (#764), and
// is judged on its own at resolution (CR 608.2b): with both chosen, one
// bullet whose creature is gone still lets the other happen.
//
// No simplification.
func init() {
	bullet := func(label string, pred effects.CardPredicate) game.ModeOption {
		opt := effects.ModeDoing(label,
			effects.TargetCreature("target creature you control", effects.And(effects.YouControl(), pred)),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Fight as One — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Fight as One — indestructible"}.Apply(ctx)
			})
		opt.Purpose = effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "indestructible"))
		opt.Purpose.Answers = game.AnswerSurvive | game.AnswerPump
		return opt
	}
	effects.Register(effects.Spec{
		OracleID:     "16f93781-6740-40c7-a727-7911beac4e74",
		Name:         "Fight as One",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			bullet("Target Human creature you control gets +1/+1 and gains indestructible until end of turn.", effects.HasSubtype("Human")),
			bullet("Target non-Human creature you control gets +1/+1 and gains indestructible until end of turn.", effects.Not(effects.HasSubtype("Human"))),
		),
	})
}
