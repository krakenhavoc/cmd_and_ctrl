package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maximus, Knight Apparent — Legendary Creature — Human Knight {3}{R}, 4/4:
//
//	"Trample
//	 When Maximus enters, you may search your library for an Equipment
//	 card with mana value 2, reveal it, put it into your hand, then
//	 shuffle.
//	 {1}, Sacrifice an artifact: You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1. The search is optional (CR 701.23b). The sacrifice may
// be any artifact you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23f0d43c-f84d-48cf-8ee7-bfda6e0f7e29",
		Name:            "Maximus, Knight Apparent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			b20TutorOnETB(
				"Maximus — search for an Equipment card with mana value 2",
				"Maximus — an Equipment card with mana value 2",
				func(c game.Card) bool { return c.HasSubtype("Equipment") && c.ManaValue() == 2 },
			),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice an artifact: You get {E}{E}.",
			Cost:    Plus(ManaCost("{1}"), game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())}),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 2},
			Effect:  Do(GetEnergy{N: 2}),
		}},
	})
}
