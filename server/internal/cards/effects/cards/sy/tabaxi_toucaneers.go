package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tabaxi Toucaneers — Creature — Cat Ranger {4}{W}, 3/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Flying is printed; myriad is the shared trigger (myriad.go, CR 702.116),
// as on Banshee of the Dread Choir.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36dd999e-53da-45c7-ae98-90f7aad78680",
		Name:            "Tabaxi Toucaneers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Tabaxi Toucaneers"),
		},
	})
}
