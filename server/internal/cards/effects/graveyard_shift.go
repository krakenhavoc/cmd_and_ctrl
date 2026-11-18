package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "da8a93c6-66e2-492d-8894-397ba9ec5a87",
		Name:         "Graveyard Shift",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInYourGraveyard(),
		Purpose:      ForTargets(Reanimates(0)),
		FlashIf: FlashIfYou("as long as there are five or more mana values among cards in your graveyard",
			func(g *game.Game, you uuid.UUID) bool { return sncFiveMVs(g, you) }),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
