package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye Gouge — Instant {B} (Born of the Gods, S82 OLD b378, #5559):
//
//	"Target creature gets -1/-1 until end of turn. If it's a Cyclops, destroy
//	 it."
//
// The shrink, then the type read as the spell resolves, so a creature that
// became a Cyclops (or stopped being one) is judged as it is now (CR 608.2b);
// the destroy happens even if the -1/-1 would have killed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89661ef9-cb42-401a-ae45-b57bb8dac033",
		Name:         "Eye Gouge",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1, Label: "Eye Gouge — -1/-1"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok || !effects.OfSubtype("Cyclops")(ctx.Game, ctx.Controller(), c) {
				return nil
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
