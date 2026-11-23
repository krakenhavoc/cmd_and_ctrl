package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Juvenile Mist Dragon — Creature — Dragon {3}{U}{U}, 4/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Confounding Clouds — When this creature enters, for each opponent, tap
//	 up to one target creature that player controls. Each of those
//	 creatures doesn't untap during its controller's next untap step."
//
// Molten Primordial's one "up to one" clause per opponent. Each pick still
// legal at resolution is tapped and held through its controller's next untap
// step (Rooftop Nuisance's TapAndFreeze).
//
// No simplification.
func init() {
	const label = "Juvenile Mist Dragon — tap up to one target creature each opponent controls; they don't untap during their controllers' next untap steps"
	enters := effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		return effects.TapAndFreeze{Targets: ids, Label: label}.Apply(ctx)
	})
	enters.TargetsFrom = effects.MoltenPrimordialClauses
	effects.Register(effects.Spec{
		OracleID:        "c6867301-296c-4840-aad4-1ad8345c091c",
		Name:            "Juvenile Mist Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
