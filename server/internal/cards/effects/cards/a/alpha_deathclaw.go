package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpha Deathclaw — Creature — Lizard Mutant {4}{B}{G}, 6/6 (Fallout,
// #5948):
//
//	"Menace, trample
//	 When this creature enters or becomes monstrous, destroy target
//	 permanent.
//	 {5}{B}{G}: Monstrosity 4. (If this creature isn't monstrous, put
//	 four +1/+1 counters on it and it becomes monstrous.)"
//
// One printed trigger over two events (OnAny: enters, becomes
// monstrous), Arbor Colossus's monstrosity line and its destroy of a
// chosen target. Any permanent is a legal target, this creature
// included; with nothing left to target the trigger is removed (CR
// 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a7d0d1f-a5ac-4db9-860c-9fcbe3fd9e71",
		Name:            "Alpha Deathclaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "trample"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{5}{B}{G}"), 4)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventETB, game.EventBecameMonstrous}, effects.Self,
				"Alpha Deathclaw — destroy target permanent", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target permanent")),
		},
	})
}
