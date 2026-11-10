package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyseer's Chariot — Artifact — Vehicle {1}{W}, 3/3 (Aetherdrift,
// #4464, #5382):
//
//	"Flying
//	 As this Vehicle enters, choose a nonland card name.
//	 Activated abilities of sources with the chosen name cost {2} more
//	 to activate.
//	 Crew 2"
//
// The name is chosen as it enters (CR 614.12) through the shared
// card-name prompt, per instance, so two Chariots can name two
// different cards and a copy names its own (CR 706.2). "A nonland card
// name" is kept where the name is read, as Phyrexian Revoker keeps it:
// a source matches only through a face whose own type line is not a
// land's (OfASourceWithTheChosenNonlandName).
//
// The tax is an activation cost increase (ActivationCostsMore, CR
// 602.2b through 601.2f): {2} more on every activated ability of a
// named source, whoever controls it — including a {T}-only ability, a
// mana ability (CR 605.3: it still has to be paid, before the mana is
// made), a loyalty ability (CR 606.4) and crew. Two Chariots naming the
// same card add {2} each. Crew 2 is the ordinary CR 702.122 ability,
// and the Chariot's own name may be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92d28545-4403-4c24-8547-efab4dde6a53",
		Name:            "Skyseer's Chariot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        ChooseCardNameAsEnters("Skyseer's Chariot — choose a nonland card name"),
		CostModifiers: []game.CostModifier{
			ActivationCostsMore(2,
				"Skyseer's Chariot — activated abilities of sources with the chosen name cost {2} more to activate.",
				OfASourceWithTheChosenNonlandName()),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Skyseer's Chariot"),
		}},
	})
}
