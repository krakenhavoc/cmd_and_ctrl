package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Filth — Creature — Incarnation {3}{B}, 2/2 (Judgment):
//
//	"Swampwalk
//	 As long as this card is in your graveyard and you control a Swamp,
//	 creatures you control have swampwalk."
//
// Brawn's graveyard anthem (CR 113.6c, #1221) for swampwalk, with the body's
// own printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "411484ac-a87d-49e8-a7ef-0cd055a872b6",
		Name:            "Filth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Static:          []game.StaticAbility{effects.IncarnationAnthem("swampwalk", "Swamp")},
	})
}
