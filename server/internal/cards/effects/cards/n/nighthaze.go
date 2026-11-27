package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nighthaze — Sorcery {B} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Target creature gains swampwalk until end of turn. (It can't be blocked
//	 as long as defending player controls a Swamp.)
//	 Draw a card."
//
// Accelerate's shape with swampwalk (an engine keyword, CR 702.14) for haste.
// Written as steps (CR 608.2c), so a target that has gone by resolution only
// skips the grant and the card is still drawn when the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0044283-7034-4d1d-876a-a40ea15b7cc6",
		Name:         "Nighthaze",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "swampwalk"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"swampwalk"}, Label: "Nighthaze — swampwalk until end of turn"}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
