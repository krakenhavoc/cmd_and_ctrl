package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expedition Healer — Creature — Kor Cleric {1}{W}, 2/2 (Zendikar Rising,
// #5508):
//
//	"Vigilance
//	 This creature has lifelink as long as you control another Cleric."
//
// Barrow Naughty's lifelink: a self-only layer 6 keyword grant gated on
// some OTHER Cleric its controller controls, read live, so a changeling
// counts and the grant ends the moment the last other Cleric leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8fbdf4b-ef64-4b82-b67f-ae5568772256",
		Name:            "Expedition Healer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			KeywordGrant(hobA9WhileYouControlAnother("Cleric"), "lifelink"),
		},
	})
}
