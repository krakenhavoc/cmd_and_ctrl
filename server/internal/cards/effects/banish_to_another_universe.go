package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banish to Another Universe — Enchantment {4}{W} (Doctor Who):
//
//	"Affinity for historic permanents (This spell costs {1} less to cast
//	 for each artifact, legendary, and/or Saga permanent you control.)
//	 When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Affinity is the shared self cost modifier over the historic predicate
// (artifact, legendary or Saga); the exile is CR 610.3's one-shot, as
// Banishing Light.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae6ef097-bd52-4bf8-81a8-2e6c643fc805",
		Name:         "Banish to Another Universe",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for historic permanents", b14Historic()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Banish to Another Universe — exile target nonland permanent an opponent controls until this enchantment leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Banish to Another Universe — the exiled card returns when it leaves the battlefield")),
				TargetPermanent("target nonland permanent an opponent controls", And(Nonland(), OpponentControls()))),
		},
	})
}
