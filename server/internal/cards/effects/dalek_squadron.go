package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dalek Squadron — Artifact Creature — Dalek {2}{B}, 3/3 (Doctor Who, #5948):
//
//	"Menace
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Menace is a printed keyword; myriad is the shared trigger (myriad.go, CR
// 702.116), as on Banshee of the Dread Choir.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18ae6be7-cfb6-42e3-a04e-99e9c1471ecd",
		Name:            "Dalek Squadron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{GrantedMyriad("Dalek Squadron")},
	})
}
