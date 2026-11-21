package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enlarge — Sorcery, {3}{G}{G}:
//
//	"Target creature gets +7/+7 and gains trample until end of turn.
//	 It must be blocked this turn if able."
//
// #1684 leftover: BoostUntilEOT, GrantKeywordUntilEOT (trample) and
// the Irresistible Prey shape (BlockRequirementUntilEOT, must-be-
// blocked). No new machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "405cd74a-6b44-4a80-b53c-ff1406d8ff45",
		Name:         "Enlarge",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 7, 7, "trample")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{
					Target:    t.ID,
					Power:     7,
					Toughness: 7,
					Label:     "Enlarge — +7/+7",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"trample"},
					Label:    "Enlarge — trample",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.BlockRequirementUntilEOT{
					Target: t.ID,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Enlarge — must be blocked this turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
