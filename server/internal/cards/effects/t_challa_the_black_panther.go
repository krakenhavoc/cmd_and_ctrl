package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// T'Challa, the Black Panther — Legendary Creature — Human Noble Hero
// {1}{G}{W}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever T'Challa enters or attacks, create a tapped Vibranium token.
//	 (It's an artifact with indestructible and "{T}: Add {C}. This mana
//	 can't be spent to cast a nonartifact spell.")
//	 Whenever you cast an artifact spell with mana value 4 or greater, put
//	 two +1/+1 counters on T'Challa."
//
// Two printed triggers, two registered. The first is one ability with two
// conditions (WhenThisEntersOrAttacks) making Dora Milaje Elite's printed
// Vibranium token tapped. The second is a cast trigger over
// And(Artifact(), ManaValueGE(4)), the mana value read on the stack with
// X included (CR 202.3e); the counters go on only if T'Challa is still on
// the battlefield (CounterOnThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae10eb15-28e7-46d1-9a2e-85763d93febf",
		Name:         "T'Challa, the Black Panther",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("T'Challa, the Black Panther — create a tapped Vibranium token",
				Do(CreateTokenAdvanced{Spec: Token(VibraniumToken()).EntersTapped(), N: 1})),
			WheneverYouCast(And(Artifact(), ManaValueGE(4)), "T'Challa, the Black Panther — two +1/+1 counters on T'Challa",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
