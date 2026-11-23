package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liquimetal Coating — Artifact {2}:
//
//	"{T}: Target permanent becomes an artifact in addition to its other types
//	 until end of turn."
//
// Liquimetal Torque's data-only layer-4 type add, on any permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4bdc551-c2eb-4a34-a3e3-b4a017c925af",
		Name:         "Liquimetal Coating",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target permanent becomes an artifact in addition to its other types until end of turn",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  liquimetalCoatingEffect,
		}},
	})
}

func liquimetalCoatingEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := effects.NewContext(g, item)
	return effects.ScopedEffectFor{
		Target:   item.Targets[0].ID,
		Mods:     []game.Mod{game.AddTypesMod("Artifact")},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Liquimetal Coating — becomes an artifact in addition to its other types",
	}.Apply(ctx)
}
