package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ammit Eternal — Creature — Zombie Crocodile Demon {2}{B}, 5/5 (Hour of
// Devastation, #5534):
//
//	"Afflict 3 (Whenever this creature becomes blocked, defending player loses
//	 3 life.)
//	 Whenever an opponent casts a spell, put a -1/-1 counter on this creature.
//	 Whenever this creature deals combat damage to a player, remove all -1/-1
//	 counters from it."
//
// Afflict is the shared trigger (CR 702.130). The counter goes on whenever any
// opponent casts a spell, once per spell, and every -1/-1 counter comes off
// when the Eternal hurts a player in combat; with none on it the removal does
// nothing (CR 609.3). A source that has left takes no counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2f6ddc2-86bc-43da-9bdd-0c2af79c62fd",
		Name:         "Ammit Eternal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Afflict(3, "Ammit Eternal"),
			effects.On(game.EventCast, effects.ByAnOpponent, "Ammit Eternal — put a -1/-1 counter on it",
				effects.PutACounterOnThis(game.CounterMinusOne)),
			effects.WheneverThisDealsCombatDamageToAPlayer("Ammit Eternal — remove all -1/-1 counters from it",
				effects.RemoveAllCountersFromThisEffect(game.CounterMinusOne)),
		},
	})
}
