package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dramatist's Puppet — Artifact Creature — Construct {4}, 2/4 (Kamigawa:
// Neon Dynasty):
//
//	"When this creature enters, for each kind of counter on target permanent,
//	 put another counter of that kind on it or remove one from it."
//
// The kinds are fixed as the ability resolves: every kind the target
// permanent has a counter of then, in a stable order. Its controller is asked
// once per kind (Plague Boiler's put-or-remove prompt, chained: the
// remaining kinds and the target ride the prompt as data, PickOption.Words
// and Carry), and that kind is then raised or lowered by one. A target that
// has left, or become illegal, stops the chain; a kind whose counter is gone
// cannot be removed and is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3473ed5-f2b2-4a41-aed9-ee37caa9d7be",
		Name:         "Dramatist's Puppet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dramatist's Puppet — for each kind of counter on target permanent, put another counter of that kind on it or remove one from it",
				effects.OldB242ForEachKindPutOrRemove()),
				effects.TargetPermanent("target permanent")),
		},
	})
}
