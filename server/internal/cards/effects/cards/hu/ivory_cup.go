package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivory Cup — Artifact {1} (Eighth Edition, #5607):
//
//	"Whenever a player casts a white spell, you may pay {1}. If you do, you
//	 gain 1 life."
//
// Wooden Sphere's trigger over white spells: the same watch on every player's
// casts, the same optional payment asked as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e2017c3-057d-4e30-bd60-f486fddbc6ca",
		Name:         "Ivory Cup",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB608CastASpellOfColor("W"),
				"Ivory Cup — you may pay {1}; if you do, you gain 1 life", effects.OldB608MayPayOneToGainOne("Ivory Cup")),
		},
	})
}
