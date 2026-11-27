package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlewand Oak — Creature — Treefolk Warrior {2}{G}, 1/3 (Lorwyn):
//
//	"Whenever a Forest you control enters, this creature gets +2/+2 until end
//	 of turn.
//	 Whenever you cast a Treefolk spell, this creature gets +2/+2 until end of
//	 turn."
//
// Two triggers, one per printed sentence. The Forest is any land of that type
// entering under its controller's control, the Oak's own turn or not; the
// Treefolk is the spell's type as cast, read through the Oak's text
// (CR 612.2). Each pump needs the Oak to still be the permanent that
// triggered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98d3d1f1-ac72-4b1e-87da-b4c1e4f03d08",
		Name:         "Battlewand Oak",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Cmdr2S14SubtypeLandEntersUnderYourControl("Forest"),
				"Battlewand Oak — this creature gets +2/+2 until end of turn", effects.S82OldB550ThisGetsPlusTwo("Battlewand Oak")),
			effects.On(game.EventCast, effects.YouCast(effects.OfSubtype("Treefolk")),
				"Battlewand Oak — this creature gets +2/+2 until end of turn", effects.S82OldB550ThisGetsPlusTwo("Battlewand Oak")),
		},
	})
}
