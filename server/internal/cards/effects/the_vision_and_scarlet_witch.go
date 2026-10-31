package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Vision and Scarlet Witch — Legendary Artifact Creature — Mutant Hero
// {2}{R}{R}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever you cast a spell, add {R} and put a +1/+1 counter on The
//	 Vision and Scarlet Witch."
//
// One printed trigger: Birgi's cast trigger over any spell, doing both
// halves in order. The {R} is added on resolution and, as the trigger
// sits above the spell, arrives before it resolves; it empties with the
// step as ordinary mana does (the card says nothing of keeping it). The
// counter goes on only while the creature is still on the battlefield
// (CounterOnThis); the mana does not depend on that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33219792-f0a7-4849-bfca-70ee2d410b9e",
		Name:            "The Vision and Scarlet Witch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "The Vision and Scarlet Witch — add {R} and put a +1/+1 counter on it",
				Do(AddMana{Produced: "{R}"}, CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
