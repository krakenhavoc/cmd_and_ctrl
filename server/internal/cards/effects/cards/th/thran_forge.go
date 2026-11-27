package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Forge — Artifact {3} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"{2}: Until end of turn, target nonartifact creature gets +1/+0 and becomes an artifact in
//	 addition to its other types."
//
// Liquimetal Coating's type change (a layer-4 mod that keeps every other type) with a +1/+0 boost, both
// until end of turn, on a creature that is not an artifact as it is chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb4414a6-27d0-4922-938b-391283535f24",
		Name:         "Thran Forge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Until end of turn, target nonartifact creature gets +1/+0 and becomes an artifact in addition to its other types.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target nonartifact creature", effects.Not(effects.Artifact())),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Thran Forge — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddTypesMod("Artifact")},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Thran Forge — becomes an artifact in addition to its other types",
				}.Apply(ctx)
			},
		}},
	})
}
