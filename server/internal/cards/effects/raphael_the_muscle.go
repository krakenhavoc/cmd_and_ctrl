package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael, the Muscle — Legendary Creature — Mutant Ninja Turtle {4}{R},
// 4/4 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Double all damage that creatures you control with counters on them
//	 would deal.
//	 When Raphael enters, create a Mutagen token.
//	 Partner—Character select"
//
// The doubling is a damage replacement (CR 701.10g) gated on the source:
// a creature Raphael's controller controls with one or more counters of
// any kind on it, combat and noncombat damage alike, to any permanent or
// player (tmcB6DoubleDamageFromCreaturesWithCounters). Raphael himself
// counts once he has a counter. The damage stays the source's own; it is
// divided or assigned first and then doubled, as the rulings say. With a
// second replacement the affected player orders them (CR 616.1), and one
// that prevents all of it leaves nothing to double.
//
// The enters trigger makes the CR 111.10v Mutagen (MutagenToken).
// Partner—Character select is a deck-construction rule the deck importer
// reads (CR 702.124i).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22008c3b-6e15-41cc-a15c-7871495191aa",
		Name:         "Raphael, the Muscle",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			tmcB6DoubleDamageFromCreaturesWithCounters("Raphael, the Muscle: double damage from a creature you control with counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Raphael, the Muscle — create a Mutagen token", Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
