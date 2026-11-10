package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forensic Gadgeteer — Creature — Vedalken Artificer Detective {2}{U},
// 2/3 (Murders at Karlov Manor, #4469):
//
//	"Whenever you cast an artifact spell, investigate. (Create a Clue
//	 token. It's an artifact with "{2}, Sacrifice this token: Draw a
//	 card.")
//	 Activated abilities of artifacts you control cost {1} less to
//	 activate. This effect can't reduce the mana in that cost to less than
//	 one mana."
//
// The trigger is Case of the Ransacked Lab's cast watcher over artifact
// spells, making a Clue (Magnifying Glass's investigate). The reduction is
// Training Grounds': an activation-scoped board reduction with the
// one-mana floor, over artifacts on the battlefield you control
// (OfAnArtifactYouControl). Triggered abilities have no activation cost to
// reduce, and an artifact card's ability that works from another zone is
// not reduced (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d5917d5-4fb3-4f37-9278-e08c1f73aae5",
		Name:         "Forensic Gadgeteer",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			KeepsOneMana(ActivationCostsLess(1,
				"Activated abilities of artifacts you control cost {1} less to activate. This effect can't reduce the mana in that cost to less than one mana.",
				OfAnArtifactYouControl())),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Artifact(), "Forensic Gadgeteer — investigate", Do(Investigate{N: 1})),
		},
	})
}
