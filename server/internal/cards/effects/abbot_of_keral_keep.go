package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abbot of Keral Keep — Creature — Human Monk {1}{R}, 2/1 (Magic Origins, #5549):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	 When this creature enters, exile the top card of your library. Until end of turn, you may play that card."
//
// Prowess is the engine keyword; the entry is Irascible Wolverine's
// impulse draw (outpostSiegeImpulse): the card is exiled and its
// controller may play it until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09be0f69-9c7f-491f-964d-bfbbc2d4251f",
		Name:            "Abbot of Keral Keep",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Abbot of Keral Keep — exile the top card of your library; until end of turn, you may play it",
				outpostSiegeImpulse),
		},
	})
}
