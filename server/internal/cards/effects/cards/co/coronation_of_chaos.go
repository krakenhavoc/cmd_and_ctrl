package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coronation of Chaos — Sorcery {2}{R} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Up to three target creatures can't block this turn. Goad them. (Until your
//	 next turn, those creatures attack each combat if able and attack a player
//	 other than you if able.)"
//
// Each target still legal as the spell resolves gets the can't-block
// restriction until end of turn (Ahn-Crop Crasher's), and all of them are goaded
// by the caster together (CR 701.15, GoadAndScheduleClear: the goad ends at the
// caster's next turn). A target that left or became illegal is skipped
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46f1618d-9c95-4a4b-9a20-7c378b811849",
		Name:         "Coronation of Chaos",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)), Answers: game.AnswerRestrict},
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			for _, id := range ids {
				if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Coronation of Chaos — can't block this turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.GoadAndScheduleClear(ctx, ids)
		},
	})
}
