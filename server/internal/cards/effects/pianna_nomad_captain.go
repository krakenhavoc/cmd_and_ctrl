package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pianna, Nomad Captain — Legendary Creature — Human Nomad {1}{W}{W},
// 2/2 (Commander Masters, #5948):
//
//	"Whenever Pianna attacks, attacking creatures get +1/+1 until end
//	 of turn."
//
// The set of attackers is fixed as the trigger resolves (CR 611.2c), so a
// creature put onto the battlefield attacking afterwards gets nothing.
// Pianna is itself attacking, so it gets the +1/+1 too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cd8f7f6-0b69-466a-8f4c-d864ff682457",
		Name:         "Pianna, Nomad Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pianna, Nomad Captain — attacking creatures get +1/+1 until end of turn",
				Do(BoostUntilEOT{Match: AttackingCreature(), Power: 1, Toughness: 1,
					Label: "Pianna, Nomad Captain — attacking creatures get +1/+1"})),
		},
	})
}
