package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Draconautics Engineer — Creature — Goblin Artificer {1}{R}, 2/2
// (Aetherdrift, #4464):
//
//	"Exhaust — {R}: Other creatures you control gain haste until end of
//	 turn. Put a +1/+1 counter on this creature. (Activate each exhaust
//	 ability only once.)
//	 Exhaust — {3}{R}: Create a 4/4 red Dinosaur Dragon creature token
//	 with flying."
//
// Two exhaust abilities, each activated once (CR 702.177a). The record is
// keyed by object and ability, so spending one leaves the other as it was
// (Greenbelt Guardian's proof).
//
// The first is Crashing Drawbridge's mass haste with "other" read as
// object identity (CR 400.7): every creature you control as it resolves
// but the Engineer that activated it, the set fixed then (CR 611.2c).
// The counter lands only while the Engineer is still that permanent
// (Pacesetter Paragon's step); one that left and came back in response
// is a new object, gets no counter, and gains haste like any other
// creature.
//
// The second makes the Dragon whether or not the Engineer is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "380d8d64-9aec-4656-b5b8-d336ebf803c8",
		Name:         "Draconautics Engineer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {R}: Other creatures you control gain haste until end of turn. Put a +1/+1 counter on this creature.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, SourceCounters: 1},
				Exhaust: true,
				Cost:    ManaCost("{R}"),
				// The counter waits for any prompt the grant asks (CR 608.2c).
				Effect: Do(
					Instruction(func(ctx *Context) error {
						return GrantKeywordUntilEOT{
							Match:    And(Creature(), YouControl(), trekB11OtherThanSourceObject(ctx.Item)),
							Keywords: []string{"haste"},
							Label:    "Draconautics Engineer — haste",
						}.Apply(ctx)
					}),
					Instruction(plusOneCounterOnThisIfStillHere),
				),
			},
			{
				Label:   "Exhaust — {3}{R}: Create a 4/4 red Dinosaur Dragon creature token with flying.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
				Exhaust: true,
				Cost:    ManaCost("{3}{R}"),
				Effect:  Do(CreateToken{Template: TokenCard("4/4 red Dinosaur Dragon with flying"), N: 1}),
			},
		},
	})
}
