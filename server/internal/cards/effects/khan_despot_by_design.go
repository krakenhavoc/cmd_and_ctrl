package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khan, Despot by Design — Legendary Creature — Human Rogue {4}{R}, 5/4
// (Star Trek, #4973):
//
//	"Trample
//	 When Khan enters, he deals 1 damage to each opponent and each
//	 creature your opponents control."
//
// The enters trigger is Tectonic Hazard's sweep dealt by Khan: one damage
// to every opponent and to every creature an opponent controls, the set
// fixed before the first point lands. Khan is the source even if he has
// left the battlefield by the time it resolves (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7ea2089-abcd-4e60-8201-93801eae5084",
		Name:            "Khan, Despot by Design",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Khan, Despot by Design — 1 damage to each opponent and each creature your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachOpponentAndEachCreatureTheyControl(NewContext(g, item), 1)
				}),
		},
	})
}
