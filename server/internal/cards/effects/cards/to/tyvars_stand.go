package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyvar's Stand — Instant {X}{G}:
//
//	"Target creature you control gets +X/+X and gains hexproof and
//	 indestructible until end of turn. (It can't be the target of
//	 spells or abilities your opponents control. Damage and effects
//	 that say "destroy" don't destroy it.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4247b667-a0a8-4995-97c7-622d20132f7d",
		Name:         "Tyvar's Stand",
		Purpose:      effects.ForTargets(effects.PumpXUntilEOT(0, "hexproof", "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			x := ctx.X()
			if err := (effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Tyvar's Stand — +X/+X"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Tyvar's Stand — hexproof and indestructible",
			}.Apply(ctx)
		},
	})
}
