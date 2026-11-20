package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blessed Spirits — Creature — Spirit {2}{W}, 2/2 (Magic Origins, #5549):
//
//	"Flying
//	 Whenever you cast an enchantment spell, put a +1/+1 counter on this creature."
//
// Flying, and Kami of Transience's cast trigger: the counter goes on as
// the trigger resolves, so it is on the stack above the enchantment
// spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fb899c1-3a2f-4b5e-a82d-fece51bdf786",
		Name:            "Blessed Spirits",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Enchantment(),
				"Blessed Spirits — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
