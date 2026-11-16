package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chittering Dispatcher — Creature — Eldrazi Drone {2}{G}, 2/3 (Modern
// Horizons 3 Commander, #5468):
//
//	"Devoid (This card has no color.)
//	 Myriad
//	 When this creature leaves the battlefield, create a 0/1 colorless
//	 Eldrazi Spawn creature token with "Sacrifice this token: Add {C}.""
//
// Myriad is the shared trigger (myriad.go): one "you may" per opponent
// other than the defending player, the copies exiled at end of combat.
// The leaves trigger is any exit, not only dying, and each myriad copy
// is a Chittering Dispatcher too, so each copy exiled at end of combat
// leaves an Eldrazi Spawn behind (the printed token, sacrifice ability
// included).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1b276b7-f1b3-4f6f-b6a2-1315d805e0f2",
		Name:            "Chittering Dispatcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Chittering Dispatcher"),
			WhenThisLeaves("Chittering Dispatcher — create a 0/1 colorless Eldrazi Spawn",
				Do(CreateToken{Template: EldraziSpawnToken(), N: 1})),
		},
	})
}
