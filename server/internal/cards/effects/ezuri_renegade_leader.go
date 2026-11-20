package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ezuri, Renegade Leader — Legendary Creature — Elf Warrior {1}{G}{G}, 2/2
// (Scars of Mirrodin, #5577):
//
//	"{G}: Regenerate another target Elf.
//	 {2}{G}{G}{G}: Elf creatures you control get +3/+3 and gain trample
//	 until end of turn."
//
// The first row is Asceticism's regeneration shield (CR 701.19a) on
// "another target Elf", an Elf permanent other than Ezuri itself. The
// second is Overrun's pump and trample over the Elf creatures its
// controller controls as it resolves (CR 611.2c: an Elf that arrives later
// gets nothing).
//
// No simplification.
func init() {
	elves := func() CardPredicate { return And(Creature(), YouControl(), Subtype("Elf")) }
	Register(Spec{
		OracleID:     "6734497c-16f0-4c4b-ba24-337333511fc6",
		Name:         "Ezuri, Renegade Leader",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{G}: Regenerate another target Elf.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{G}"),
				Targets: Another(TargetPermanent("another target Elf", Subtype("Elf"))),
				Effect:  regenerateTheTargetPermanent,
			},
			{
				Label:   "{2}{G}{G}{G}: Elf creatures you control get +3/+3 and gain trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{2}{G}{G}{G}"),
				Effect: Do(Instruction(func(ctx *Context) error {
					if err := (BoostUntilEOT{
						Match: elves(), Power: 3, Toughness: 3,
						Label: "Ezuri, Renegade Leader — +3/+3",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{
						Match: elves(), Keywords: []string{"trample"},
						Label: "Ezuri, Renegade Leader — trample",
					}.Apply(ctx)
				})),
			},
		},
	})
}
