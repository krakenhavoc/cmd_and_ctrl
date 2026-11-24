package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weigh Down — Sorcery {B} (Kaldheim):
//
//	"As an additional cost to cast this spell, exile a creature card from
//	 your graveyard.
//	 Target creature gets -3/-3 until end of turn."
//
// The additional cost is the mandatory exile-cards slot of a cast
// (ExileFromGraveyardCost): the caster names a creature card in their own
// graveyard, so the spell cannot be cast with none (CR 601.2b, 601.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "51f1ceb9-2305-4a4e-b422-7dd7f8deef98",
		Name:           "Weigh Down",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool { return c.IsCreature() }),
		Targets:        effects.TargetCreature("target creature"),
		Purpose:        effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BoostTargetUntilEOT("Weigh Down — -3/-3 until end of turn", -3, -3)(ctx.Game, item)
		},
	})
}
