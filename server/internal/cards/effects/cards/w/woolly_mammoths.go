package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woolly Mammoths — Creature — Elephant {1}{G}{G}, 3/2 (Ice Age, #5636):
//
//	"This creature has trample as long as you control a snow land."
//
// Wingrattle Scarecrow's shape: a layer-6 keyword grant to itself while
// its controller controls a land with the snow supertype (CR 205.4g),
// read off each land's effective type line, so a land that stops being
// snow stops counting.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e5f1088-a5ab-48ec-be27-c88eb164cf63",
		Name:         "Woolly Mammoths",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SelfWhileYouControlA(effects.And(effects.Land(), effects.Snow())), "trample"),
		},
	})
}
