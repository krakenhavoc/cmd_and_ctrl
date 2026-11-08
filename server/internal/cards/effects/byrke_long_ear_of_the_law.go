package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Byrke, Long Ear of the Law — Legendary Creature — Rabbit Soldier
// {4}{G}{W}, 4/4 (Bloomburrow, #4467):
//
//	"Vigilance
//	 When Byrke enters, put a +1/+1 counter on each of up to two target
//	 creatures.
//	 Whenever a creature you control with a +1/+1 counter on it attacks,
//	 double the number of +1/+1 counters on it."
//
// The entry is Rishkar, Peema Renegade's "up to two target creatures",
// one counter on each one still legal (CR 608.2b). The attack trigger is
// the per-attacker EventAttack one (Wolverine, Claws Out's shape): the
// creature is judged as it attacks, so one with no +1/+1 counter does not
// trigger it. The doubling is Primordial Hydra's (b23DoublePlusOneCountersOn):
// as many +1/+1 counters again as it has as the trigger resolves (CR
// 701.10e), placed through AddCounter so a counter doubler applies. A
// creature that left the battlefield, or lost its counters, gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e9c6e4e6-0c3f-4048-93e6-161a1221ea36",
		Name:            "Byrke, Long Ear of the Law",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Byrke, Long Ear of the Law — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
			On(game.EventAttack, byrkeCounteredCreatureAttacks,
				"Byrke, Long Ear of the Law — double the +1/+1 counters on that creature",
				func(g *game.Game, item *game.StackItem) error {
					return b23DoublePlusOneCountersOn(NewContext(g, item), item.Trigger.Event.CardID)
				}),
		},
	})
}

// byrkeCounteredCreatureAttacks is "a creature you control with a +1/+1
// counter on it attacks", judged as the attacker is declared.
func byrkeCounteredCreatureAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller && c.Counters[game.CounterPlusOne] > 0
}
