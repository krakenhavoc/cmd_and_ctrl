package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magnigoth Treefolk — Creature — Treefolk {4}{G}, 2/6 (Planeshift, S82 OLD b642, #5615):
//
//	"Domain — For each basic land type among lands you control, this creature has
//	 landwalk of that type. (It can't be blocked as long as defending player
//	 controls a land of that type.)"
//
// A layer 6 static (CR 613.1f) that grants one landwalk word per basic land type
// its controller controls (CR 305.6), each once however many lands carry it. It
// reads the layered lands, so a land that gained a type counts and one that lost
// its types does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b153e4c4-6557-43a6-8404-e4450f879945",
		Name:         "Magnigoth Treefolk",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB642DomainLandwalk("Domain — For each basic land type among lands you control, this creature has landwalk of that type.")},
	})
}
