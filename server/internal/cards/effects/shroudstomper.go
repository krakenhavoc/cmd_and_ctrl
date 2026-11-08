package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shroudstomper — Creature — Elemental {3}{W}{W}{B}{B}, 5/5 (Duskmourn:
// House of Horror, #4466):
//
//	"Deathtouch
//	 Whenever this creature enters or attacks, each opponent loses 2
//	 life. You gain 2 life and draw a card."
//
// The Overlords' enters-or-attacks trigger (WhenThisEntersOrAttacks)
// with Artificial Anomaly's drain: every opponent loses 2, then you gain
// a flat 2 (not the total lost), then you draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36415fe3-6bf3-4fec-909c-6c16584c44eb",
		Name:            "Shroudstomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Shroudstomper — each opponent loses 2 life; you gain 2 life and draw a card",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}, DrawCards{N: 1})),
		},
	})
}
