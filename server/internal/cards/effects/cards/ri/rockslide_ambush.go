package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rockslide Ambush — Sorcery {1}{R} (Portal Three Kingdoms, S82 #5622):
//
//	"Rockslide Ambush deals damage to target creature equal to the number of
//	 Mountains you control."
//
// The Mountains (lands with the subtype, post-layer, so a land made a Mountain
// counts) are counted as the spell resolves (Downhill Charge's count), and the
// damage is dealt by the spell. With the target illegal the spell does nothing
// (CR 608.2b); with no Mountain it deals 0 damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13958d6f-25f9-4a6c-a16c-e7a20e1b9117",
		Name:         "Rockslide Ambush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := effects.B08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain"))
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		},
	})
}
