package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadly Allure — Sorcery {B} (Dark Ascension, #5571):
//
//	"Target creature gains deathtouch until end of turn and must be
//	 blocked this turn if able.
//	 Flashback {G} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Compelled Duel's block requirement (BlockRequirementMustBeBlocked)
// beside a deathtouch grant, both until end of turn. Flashback is the
// shared alternative cost (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f9cdaecd-9621-4a9a-888e-748ecfe551fd",
		Name:             "Deadly Allure",
		Completeness:     effects.CompletenessFull,
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
		Targets:          effects.TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{G}")},
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Deadly Allure — deathtouch"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BlockRequirementUntilEOT{Target: id, Kind: game.BlockRequirementMustBeBlocked, Label: "Deadly Allure — must be blocked this turn"}.Apply(ctx)
			}),
		),
	})
}
