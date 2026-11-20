package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herald of the Host — Creature — Angel {3}{W}{W}, 4/4 (Commander 2015,
// #5546):
//
//	"Flying, vigilance
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Myriad is the shared trigger (myriad.go, CR 702.116); flying and
// vigilance are printed. The copies enter tapped and attacking and were
// never declared as attackers (CR 508.4), so vigilance does not untap them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4371d5bc-1760-477b-822e-6d4999fb0a6b",
		Name:            "Herald of the Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{GrantedMyriad("Herald of the Host")},
	})
}
