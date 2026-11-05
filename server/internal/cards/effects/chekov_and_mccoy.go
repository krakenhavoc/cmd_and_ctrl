package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chekov and McCoy — {3}{G}{U} Legendary Creature — Human Pilot Doctor
// 3/3 (Star Trek Commander, #4974):
//
//	"Whenever you put one or more +1/+1 counters on Chekov and McCoy, you
//	 may draw that many cards. Do this only once each turn."
//
// Terrasymbiosis's ability, aimed at itself: onceEachTurnMayDrawThatMany
// with the creature the counters went on being this one. One trigger per
// placement of +1/+1 counters you put on it (CR 603.2c); "that many" is
// the counters that placement put, after replacements, read as it
// resolves. "Do this only once each turn" limits the draw: the first
// "yes" draws and uses up the turn; a "no" does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "058a5b8b-08bc-4021-8a8b-8e149941e775",
		Name:         "Chekov and McCoy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			onceEachTurnMayDrawThatMany(trekB46IsThis,
				"Chekov and McCoy — draw that many cards",
				"Chekov and McCoy — cards drawn this turn",
				"Chekov and McCoy — draw that many cards?"),
		},
	})
}
