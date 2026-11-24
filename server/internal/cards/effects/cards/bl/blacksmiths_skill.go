package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blacksmith's Skill — Instant {W} (Modern Horizons 2, #5501):
//
//	"Target permanent gains hexproof and indestructible until end of turn.
//	 If it's an artifact creature, it gets +2/+2 until end of turn."
//
// The "if" is read as the spell resolves, off the permanent's effective
// types (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c5b308c-07d5-422a-8b45-b8b19cb967ca",
		Name:         "Blacksmith's Skill",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof", "indestructible"},
				Label: "Blacksmith's Skill — hexproof and indestructible"}).Apply(ctx); err != nil {
				return err
			}
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsArtifact() && c.IsCreature() {
				return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Blacksmith's Skill — +2/+2"}.Apply(ctx)
			}
			return nil
		},
	})
}
