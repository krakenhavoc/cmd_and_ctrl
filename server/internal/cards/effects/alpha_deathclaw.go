package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "0a7d0d1f-a5ac-4db9-860c-9fcbe3fd9e71",
		Name:            "Alpha Deathclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "trample"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{5}{B}{G}"), 4)},
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventETB, game.EventBecameMonstrous}, Self,
				"Alpha Deathclaw — destroy target permanent", destroyChosenPermanent),
				TargetPermanent("target permanent")),
		},
	})
}
