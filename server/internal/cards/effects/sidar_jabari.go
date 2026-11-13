package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sidar Jabari — Legendary Creature — Human Knight {3}{W}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 Whenever Sidar Jabari attacks, tap target creature defending player
//	 controls."
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). The attack trigger is Distracting Geist's
// (distractingGeistTap, which The Wasp shares): the defending player is
// read off Sidar Jabari's own attack, so in a multiplayer game only a
// creature of the player he attacks can be chosen (CR 506.2a). The tap
// resolves in the declare attackers step, before blockers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb637dfd-4022-43a4-8632-77e614436f38",
		Name:            "Sidar Jabari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered: []game.TriggeredAbility{
			Flanking("Sidar Jabari"),
			distractingGeistTap("Sidar Jabari — tap target creature defending player controls"),
		},
	})
}
