package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspirited Vanguard — Creature — Human Soldier {4}{G}, 3/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"Whenever this creature enters or attacks, it endures 2. (Put two
//	 +1/+1 counters on it or create a 2/2 white Spirit creature token.)"
//
// One ability with two trigger events (WhenThisEntersOrAttacks, Sun
// Titan's). Endure is endure_n.go (CR 701.63a): a Vanguard that has left
// the battlefield by the time the trigger resolves can't take the
// counters, so the Spirit is created.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db756a16-3db6-4bde-bbc2-ecee3ee9d9d3",
		Name:         "Inspirited Vanguard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Inspirited Vanguard — it endures 2", ThisEndures("Inspirited Vanguard", 2)),
		},
	})
}
