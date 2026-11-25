package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siren's Ruse — Instant {1}{U} (Ixalan):
//
//	"Exile target creature you control, then return that card to the
//	 battlefield under its owner's control. If a Pirate was exiled this way,
//	 draw a card."
//
// Felidar Guardian's blink on a creature you control. Whether it is a Pirate is
// read as the spell resolves, before it leaves, through the layers (a changeling
// counts); the draw follows only if the exile happened, so a target that has
// become illegal does nothing (CR 608.2b) and a commander that goes to the
// command zone instead was not exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58a55d14-b03b-4648-b168-ad35ff088b19",
		Name:         "Siren's Ruse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			pirate := c.HasSubtypeWord(ctx.TextSubtype("Pirate"))
			if err := (effects.Flicker{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !pirate {
				return nil
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneCommand {
				return nil
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
