package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Laser Screwdriver — Artifact {3} (Doctor Who, #5948):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Tap target artifact.
//	 {2}, {T}: Surveil 1.
//	 {3}, {T}: Goad target creature."
//
// Four abilities that each tap the Screwdriver, so only one can be paid per
// untap. The mana ability is the five-colour pipe (any colour, no commander
// narrowing); tapping uses the shared tap-the-targets body; goad is the
// shared GoadTarget (CR 701.15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9602b5f7-65f9-4447-815e-f60d3e4f7290",
		Name:         "Laser Screwdriver",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Tap target artifact.",
				Cost:    game.AbilityCost{Mana: "{1}", Tap: true},
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect:  tapEachLegalTarget,
			},
			{
				Label:   "{2}, {T}: Surveil 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    game.AbilityCost{Mana: "{2}", Tap: true},
				Effect:  Do(Surveil{N: 1}),
			},
			{
				Label:   "{3}, {T}: Goad target creature.",
				Cost:    game.AbilityCost{Mana: "{3}", Tap: true},
				Targets: TargetCreature("target creature"),
				Effect:  GoadTarget,
			},
		},
	})
}
