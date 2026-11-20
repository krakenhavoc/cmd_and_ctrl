package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rafiq of the Many — Legendary Creature — Human Knight {1}{G}{W}{U}, 3/3
// (Shards of Alara, #5588):
//
//	"Exalted
//	 Whenever a creature you control attacks alone, it gains double
//	 strike until end of turn."
//
// Exalted is the canonical keyword (CR 702.83a); the double strike is the
// shared alaB1LoneAttackerGains row, Peggy Carter's attacks-alone trigger
// (CR 506.5) with a keyword grant. The two triggers are separate, so the
// lone attacker gets both the +1/+1 and double strike (CR 702.4a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c6e17443-2379-419d-8c32-13ea9bf52993",
		Name:            "Rafiq of the Many",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Triggered: []game.TriggeredAbility{
			alaB1LoneAttackerGains("Rafiq of the Many", "double strike"),
		},
	})
}
