package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surrounded by Orcs — Sorcery {3}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Amass Orcs 3, then target player mills X cards, where X is the amassed
//	 Army's power."
//
// CR 701.47c: "the Army you amassed" is the Army the action chose, handed to the
// mill as the amass tail, so its power is read after the counters are on it,
// through the layers. With no Army to choose and none made (CR 701.47b) the
// player mills nothing. The target player is chosen as the spell is cast and
// checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65c5e120-7a3d-475f-9bb2-8ddb5062e4ba",
		Name:         "Surrounded by Orcs",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 3},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Amass{Subtype: "Orc", N: 3, ThenTail: surroundedByOrcsMill}.Apply(ctx)
		},
	})
}

// surroundedByOrcsMill is the mill after the amass. An on-disk key: never
// renamed, never reused.
var surroundedByOrcsMill = effects.AmassTail("then/surrounded-by-orcs-mill", func(ctx *effects.Context, army uuid.UUID) error {
	if army == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(army)
	if !ok || c.CurrentPower() <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.MillCards{Player: t.ID, N: c.CurrentPower()}.Apply(ctx)
		}
	}
	return nil
})
