package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire's Zeal — Instant {W} (Ixalan):
//
//	"Target creature gets +2/+2 until end of turn. If it's a Vampire, it
//	 gains first strike until end of turn."
//
// One target, re-checked as the spell resolves (CR 608.2b). "It's a Vampire"
// is read as the spell resolves, through the layers, and the keyword rides
// the same record as the pump on that object.
//
// No simplification.
func init() {
	const label = "Vampire's Zeal — target creature gets +2/+2 and, if it's a Vampire, first strike until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "cc7dcb89-534b-4e86-abd2-8362855e2f66",
		Name:         "Vampire's Zeal",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			mods := []game.Mod{game.ModifyPTMod(2, 2)}
			if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.HasSubtypeWord("Vampire") {
				mods = append(mods, game.AddKeywordsMod("first strike"))
			}
			return effects.UntilEndOfTurn(ctx, t.ID, nil, label, mods...)
		},
	})
}
