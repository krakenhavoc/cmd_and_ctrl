package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jump Scare — Instant {W} (Duskmourn: House of Horror, #4466):
//
//	"Until end of turn, target creature gets +2/+2, gains flying, and
//	 becomes a Horror enchantment creature in addition to its other
//	 types."
//
// One until-end-of-turn effect with four parts, each applied in its own
// layer (types, abilities, then the boost; CR 613.1), locked to the
// target as the spell resolves and ending in cleanup (CR 514.2). The
// Horror subtype is read through any text change (TextSubtypes). A target
// gone by resolution leaves the spell with no legal target, so it does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "783aab37-3ec3-478f-a4fa-cccdfffbae18",
		Name:         "Jump Scare",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Jump Scare — +2/+2, flying, Horror enchantment creature",
				game.ModifyPTMod(2, 2),
				game.AddKeywordsMod("flying"),
				game.AddTypesMod("Enchantment"),
				game.AddSubtypesMod(ctx.TextSubtypes("Horror")...))
		},
	})
}
