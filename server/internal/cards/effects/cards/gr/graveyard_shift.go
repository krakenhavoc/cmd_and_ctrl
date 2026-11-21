package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Graveyard Shift — Sorcery {4}{B} (Streets of New Capenna):
//
//	"This spell has flash as long as there are five or more mana values
//	 among cards in your graveyard.
//	 Return target creature card from your graveyard to the battlefield."
//
// The flash is a timing permission read as the cast begins (Spec.FlashIf,
// CR 601.3), the way Colossal Rattlewurm's is; the spell is not in the
// graveyard it is counting. The body is Zombify's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da8a93c6-66e2-492d-8894-397ba9ec5a87",
		Name:         "Graveyard Shift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		FlashIf: effects.FlashIfYou("as long as there are five or more mana values among cards in your graveyard",
			func(g *game.Game, you uuid.UUID) bool { return effects.SncFiveMVs(g, you) }),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
