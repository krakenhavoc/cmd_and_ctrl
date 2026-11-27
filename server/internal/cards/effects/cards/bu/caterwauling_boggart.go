package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caterwauling Boggart — Creature — Goblin Shaman {3}{R}, 2/2 (Lorwyn, S82 OLD,
// #5593):
//
//	"Goblins you control and Elementals you control have menace."
//
// A layer-6 tribal grant over the Goblins and Elementals its controller controls
// (TribalKeywordGrant, any-of, changelings included, the Boggart itself too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab3f4455-7329-4145-b9c4-93eef38bf327",
		Name:         "Caterwauling Boggart",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Goblin", "Elemental"}, YoursOnly: true}, "menace"),
		},
	})
}
