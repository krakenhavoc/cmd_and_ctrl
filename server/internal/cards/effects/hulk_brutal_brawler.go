package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hulk, Brutal Brawler — Legendary Creature — Gamma Berserker Hero
// {2}{G}{G}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Hulk attacks each combat if able.
//	 Whenever Hulk attacks, put a +1/+1 counter on each other creature
//	 you control."
//
// The requirement is AttacksEachCombat (CR 508.1d). The counters go on
// the other creatures its controller controls as the trigger resolves,
// one placement each (putAPlusOneCounterOnEachYouControl); Hulk himself
// is excluded by identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a953b71-2994-4c08-9602-b2bbce5b5b4f",
		Name:         "Hulk, Brutal Brawler",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hulk — put a +1/+1 counter on each other creature you control", msc10CounterOnEachOtherCreature),
		},
	})
}

// msc10CounterOnEachOtherCreature is "put a +1/+1 counter on each other
// creature you control" for the resolving item's source.
func msc10CounterOnEachOtherCreature(g *game.Game, item *game.StackItem) error {
	return putAPlusOneCounterOnEachYouControl(And(Creature(), OtherThan(item.SourceCardID)))(g, item)
}
