package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teyo, Lightshield Expert — Legendary Creature — Human Cleric {1}{W}, 1/1
// (Reality Fracture, tracker #2795):
//
//	"Flash
//	 When Teyo enters, target permanent you control gains hexproof until
//	 end of turn. Put a +1/+1 counter on it if it's a creature. Put a
//	 loyalty counter on it if it's a planeswalker. (It can't be the target
//	 of spells or abilities your opponents control.)"
//
// Same body as Teyo, Diamondblade Mage (teyoEntersTrigger), with hexproof.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77a9fac6-3c26-4456-9e0b-f5839a3dcae2",
		Name:            "Teyo, Lightshield Expert",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{teyoEntersTrigger("Teyo, Lightshield Expert — target permanent you control gains hexproof until end of turn", "hexproof")},
	})
}
