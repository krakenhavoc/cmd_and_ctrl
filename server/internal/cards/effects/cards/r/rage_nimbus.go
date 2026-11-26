package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rage Nimbus — {2}{R} Creature — Elemental 5/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Defender, flying
//	{1}{R}: Target creature attacks this turn if able.
//
// Into the Fray's requirement, as an activated ability.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab5b49d4-e648-43f1-bcc0-7dd4763316ee",
		Name:            "Rage Nimbus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Target creature attacks this turn if able.",
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Rage Nimbus — attacks this turn if able",
				}.Apply(ctx)
			},
		}},
	})
}
