package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quarry Hauler — Creature — Camel {3}{G}, 4/3 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, for each kind of counter on target permanent,
//	 put another counter of that kind on it or remove one from it."
//
// Dramatist's Puppet's enters trigger word for word, so the shared
// OldB242ForEachKindPutOrRemove runs both. The target may be any permanent;
// its kinds are fixed as the ability resolves and its controller is asked once
// per kind.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a55bdb73-e17a-414d-b27f-dc53af2f345c",
		Name:         "Quarry Hauler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Quarry Hauler — for each kind of counter on target permanent, put another counter of that kind on it or remove one from it",
				effects.OldB242ForEachKindPutOrRemove()),
				effects.TargetPermanent("target permanent")),
		},
	})
}
