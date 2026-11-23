package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "09be0f69-9c7f-491f-964d-bfbbc2d4251f",
		Name:            "Abbot of Keral Keep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abbot of Keral Keep — exile the top card of your library; until end of turn, you may play it",
				effects.OutpostSiegeImpulse),
		},
	})
}
