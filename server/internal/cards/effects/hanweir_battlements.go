package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hanweir Battlements — Land:
//
//	"{T}: Add {C}.
//	 {R}, {T}: Target creature gains haste until end of turn.
//	 {3}{R}{R}, {T}: If you both own and control this land and a
//	 creature named Hanweir Garrison, exile them, then meld them into
//	 Hanweir, the Writhing Township."
//
// The third ability is the meld ability (CR 701.42a, 712.5b; ADR 0145,
// #2699), at instant speed: its condition is checked as it resolves
// (MeldWith), both are exiled together and come back as Hanweir, the
// Writhing Township, a new object — so a Garrison melded mid-combat
// leaves combat, and the Township that enters is not attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e735ba6-7fd1-4d12-b20c-21525dc1e2b5",
		Name:         "Hanweir Battlements",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}, {T}: Target creature gains haste until end of turn.",
				Cost:    Plus(ManaCost("{R}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  ebTargetCreatureGainsUntilEOT("haste", "Hanweir Battlements — haste until end of turn"),
			},
			{
				Label:   "{3}{R}{R}, {T}: If you both own and control this land and a creature named Hanweir Garrison, exile them, then meld them into Hanweir, the Writhing Township.",
				Cost:    Plus(ManaCost("{3}{R}{R}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(MeldWith{Partner: "Hanweir Garrison", PartnerType: "creature"}),
			},
		},
	})
}
