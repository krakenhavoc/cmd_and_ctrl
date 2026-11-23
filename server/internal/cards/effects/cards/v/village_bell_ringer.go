package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Village Bell-Ringer — Creature — Human Scout {2}{W}, 1/4 (EDHREC
// rank 3646):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 When this creature enters, untap all creatures you control."
//
// The combat trick on a body, and the Splinter Twin-style combo
// piece. Flash rides PrintedKeywords, which the cast path reads for
// a card in hand; the entry trigger untaps every tapped creature the
// controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7b50175-b473-452a-bb30-19a71c42e649",
		Name:            "Village Bell-Ringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Village Bell-Ringer — untap all creatures you control", effects.B34UntapAllCreaturesYouControl),
		},
	})
}
