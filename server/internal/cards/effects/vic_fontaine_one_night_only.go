package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vic Fontaine, One Night Only — Legendary Artifact Creature — Illusion
// Performer {1}{R}{W}, 3/3 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, one or two
//	 target creatures each get +2/+0 and gain lifelink until end of turn."
//
// The second-card trigger fires once a turn (trekB41SeekOut). The one or
// two targets are chosen as it goes on the stack; each still legal as it
// resolves gets the pump and the lifelink.
//
// No simplification.
func init() {
	const label = "Vic Fontaine, One Night Only — one or two target creatures each get +2/+0 and gain lifelink"
	Register(Spec{
		OracleID:     "1d6f4573-d2f6-4771-a748-291846ef14b9",
		Name:         "Vic Fontaine, One Night Only",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(trekB41SeekOut(label, trekB41PumpAndGrantEachTarget(2, 0, label, "lifelink")),
				TargetCreature("one or two target creatures").WithCount(1, 2)),
		},
	})
}
