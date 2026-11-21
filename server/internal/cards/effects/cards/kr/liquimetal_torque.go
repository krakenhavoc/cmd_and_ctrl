package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liquimetal Torque — Artifact {2}:
//
//	"{T}: Add {C}.
//	 {T}: Target nonland permanent becomes an artifact in addition to
//	 its other types until end of turn."
//
// Two ordinary {T} abilities sharing one tap symbol, like the Signet
// cycle's own two-ability rocks. The type grant is a data-only
// ScopedEffectFor record (ADR 0041 phase 3 — no closure-bearing
// scoped static) with AddTypesMod, until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7d4b7dd-fbb1-4ca3-875f-ef13a95e66ad",
		Name:         "Liquimetal Torque",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target nonland permanent becomes an artifact in addition to its other types until end of turn",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect:  liquimetalTorqueEffect,
		}},
	})
}

// liquimetalTorqueEffect is the resolution body: a layer-4 type-add
// until end of turn, pinned to the target's instance and entry stamp
// by ScopedEffectFor. Package-level so the stack item captures
// nothing.
func liquimetalTorqueEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := effects.NewContext(g, item)
	return effects.ScopedEffectFor{
		Target:   item.Targets[0].ID,
		Mods:     []game.Mod{game.AddTypesMod("Artifact")},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Liquimetal Torque — becomes an artifact in addition to its other types",
	}.Apply(ctx)
}
