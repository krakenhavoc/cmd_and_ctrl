package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herald of War — Creature — Angel {3}{W}{W}, 3/3:
//
//	"Flying
//	 Whenever this creature attacks, put a +1/+1 counter on it.
//	 Angel spells and Human spells you cast cost {1} less to cast for
//	 each +1/+1 counter on this creature."
//
// The reduction counts every +1/+1 counter on Herald as the spell is
// priced and applies to generic mana only, after any increase
// (CR 601.2f).
func init() {
	Register(Spec{
		OracleID:        "1228bd6f-565d-4ad5-bd24-e5aa1ead5dee",
		Name:            "Herald of War",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLessEach(CountersOnSource(game.CounterPlusOne),
				"Angel spells and Human spells you cast cost {1} less to cast for each +1/+1 counter on this creature.",
				YourSpell(), SpellOfEitherCreatureType("Angel", "Human")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, Self, "Herald of War — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
