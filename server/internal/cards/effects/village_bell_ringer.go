package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "d7b50175-b473-452a-bb30-19a71c42e649",
		Name:            "Village Bell-Ringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Village Bell-Ringer — untap all creatures you control", b34UntapAllCreaturesYouControl)
			},
		}},
	})
}
