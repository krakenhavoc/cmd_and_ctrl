package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pterodon Knight — Creature — Human Knight {3}{W}, 3/3 (Ixalan, #5532):
//
//	"This creature has flying as long as you control a Dinosaur."
//
// Watchwing Scarecrow's self grant (layer 6, CR 613.1f), read off the
// board as it is: any Dinosaur you control counts, and the Knight loses
// flying the moment the last one leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8feb7149-a4dc-491a-a0b4-d2e8b85a97ce",
		Name:         "Pterodon Knight",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.Subtype("Dinosaur")), "flying"),
		},
	})
}
