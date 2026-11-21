package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltered Thicket — Land — Mountain Forest (EDHREC rank 4200):
//
//	"({T}: Add {R} or {G}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The red-green member of the Amonkhet cycling-land cycle — same
// shape as Polluted Mire and Canyon Slough.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "db8d8643-3d0b-4f20-bf53-f4cd26a0e8df",
		Name:          "Sheltered Thicket",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "G")},
	})
}
