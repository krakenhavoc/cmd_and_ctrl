package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liquimetal Coating — Artifact {2}:
//
//	"{T}: Target permanent becomes an artifact in addition to its other types
//	 until end of turn."
//
// Liquimetal Torque's data-only layer-4 type add, on any permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4bdc551-c2eb-4a34-a3e3-b4a017c925af",
		Name:         "Liquimetal Coating",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target permanent becomes an artifact in addition to its other types until end of turn",
			Cost:    TapCost(),
			Targets: TargetPermanent("target permanent"),
			Effect:  liquimetalCoatingEffect,
		}},
	})
}

func liquimetalCoatingEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	return ScopedEffectFor{
		Target:   item.Targets[0].ID,
		Mods:     []game.Mod{game.AddTypesMod("Artifact")},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Liquimetal Coating — becomes an artifact in addition to its other types",
	}.Apply(ctx)
}
