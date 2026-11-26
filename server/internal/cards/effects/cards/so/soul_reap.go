package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Reap — Sorcery {1}{B}:
//
//	"Destroy target nongreen creature. Its controller loses 3 life if you've
//	 cast another black spell this turn."
//
// Doom Blade's destroy, then the life loss, to the creature's controller as it
// was before the destroy (CR 608.2h), if the caster cast another black spell
// this turn (Soul Reap itself is not "another"; the turn's record of each
// spell's colours, game.CastAnotherSpellOfColorThisTurnForEffect). The loss is
// read as the spell resolves, and happens whether or not the creature was
// destroyed (a regenerated or indestructible creature still costs its
// controller), but not if the spell fizzled.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d65f701d-31a1-4275-a580-3e1799ebf8b9",
		Name:         "Soul Reap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nongreen creature", effects.Not(effects.OfColor("G"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if !effects.S82OldB526CastAnotherOfColor(ctx, item.Controller, item.SourceCardID, "B") {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim.Controller, -3)
		},
	})
}
