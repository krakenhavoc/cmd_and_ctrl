package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Condemnation — Creature — Angel {2}{W}{W}, 3/3:
//
//	"Flying, vigilance
//	 {2}{W}, {T}: Exile another target creature. Return that card to the
//	 battlefield under its owner's control at the beginning of the next
//	 end step.
//	 {2}{W}, {T}, Exert this creature: Exile another target creature
//	 until this creature leaves the battlefield. (An exerted creature
//	 won't untap during your next untap step.)"
//
// The first ability is a delayed blink (exileTargetsThenScheduleReturn).
// The second is the "until" exile keyed to the Angel as the object that
// activated it (CR 610.3): if the Angel has left the battlefield by the
// time the ability resolves, the creature isn't exiled at all. It pays
// an exert as well as the {T} (ADR 0130 §4, CR 701.43a); vigilance does
// not help, since the cost taps it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2317a33f-665e-4cbc-bb5c-16b912ac2eef",
		Name:            "Angel of Condemnation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{W}, {T}: Exile another target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
				Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
				Targets: Another(TargetCreature("another target creature")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileTargetsThenScheduleReturn(NewContext(g, item),
						"Angel of Condemnation — return the exiled creature")
				},
			},
			{
				Label:   "{2}{W}, {T}, Exert this creature: Exile another target creature until this creature leaves the battlefield.",
				Cost:    Plus(ManaCost("{2}{W}"), TapCost(), ExertThis()),
				Targets: Another(TargetCreature("another target creature")),
				Effect:  exileChosenTargetUntilThisLeaves("Angel of Condemnation — returns when the Angel leaves the battlefield"),
			},
		},
	})
}
