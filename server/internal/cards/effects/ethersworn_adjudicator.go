package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethersworn Adjudicator — Artifact Creature — Vedalken Knight {4}{U}, 4/4:
//
//	"Flying
//	 {1}{W}{B}, {T}: Destroy target creature or enchantment.
//	 {2}{U}: Untap this creature."
//
// The destroy is Mortify's, without "can't be regenerated" (the card does not
// print it), so a regeneration shield still saves the target. It is chosen as
// the ability is announced, and an illegal target as it resolves does
// nothing (CR 608.2b). Both costs are ordinary: the first needs the
// Adjudicator untapped and free of summoning sickness (CR 302.6), the second
// is mana only and untaps it again (Depth Charge Colossus' shape), so it can
// destroy something once for each {1}{W}{B}{2}{U} paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca2a7669-4932-4a78-8a6c-8558b7c05ba5",
		Name:            "Ethersworn Adjudicator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{W}{B}, {T}: Destroy target creature or enchantment.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{1}{W}{B}"), TapCost()),
				Targets: TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
				Effect:  destroyChosenPermanent,
			},
			{
				UntapSelf: true,
				Label:     "{2}{U}: Untap this creature.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      ManaCost("{2}{U}"),
				Effect:    b31UntapSelf,
			},
		},
	})
}
