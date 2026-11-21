package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Incite — Instant {R} (Magic 2011, #5579):
//
//	"Target creature becomes red until end of turn and attacks this turn
//	 if able."
//
// Cerulean Wisps' layer 5 colour change, which replaces the creature's
// colours rather than adding to them (CR 105.3), with Imps' Taunt's attack
// requirement (checked as attackers are declared, CR 508.1d), both until
// end of turn. The requirement binds only if the creature can attack this
// turn; nothing happens if it can't. "Red" is a word of the card's text
// (CR 612.2). A target gone in response means the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "304da0e0-9852-4ba5-a896-3c6bd703a64c",
		Name:         "Incite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			if err := effects.UntilEndOfTurn(ctx, id, nil, "Incite — becomes red until end of turn",
				game.SetColorsMod(ctx.TextColors("R")...)); err != nil {
				return err
			}
			return effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Incite — attacks this turn if able",
			}.Apply(ctx)
		})),
	})
}
