package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Werefox Bodyguard — Creature — Elf Fox Knight {1}{W}{W}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"Flash
//	 When this creature enters, exile up to one other target non-Fox
//	 creature until this creature leaves the battlefield.
//	 {1}{W}, Sacrifice this creature: You gain 2 life."
//
// The entry is All-Fates Stalker's exile (exileChosenTargetUntilThisLeaves,
// CR 610.3) over any creature that is neither a Fox nor the Bodyguard
// itself, whoever controls it; "up to one" lets it enter with no target.
// If the Bodyguard has left before the trigger resolves, nothing is exiled
// (CR 610.3c), and sacrificing it later returns the exiled creature. The
// sacrifice is Gingerbrute's life row without the tap. Flash rides
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d5ee2ced-29f4-430f-962e-2f930b92624c",
		Name:            "Werefox Bodyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Werefox Bodyguard — exile up to one other target non-Fox creature until this creature leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Werefox Bodyguard — the exiled creature returns when Werefox Bodyguard leaves the battlefield")),
				Another(TargetCreature("up to one other target non-Fox creature", Not(OfSubtype("Fox"))).WithCount(0, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{W}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
