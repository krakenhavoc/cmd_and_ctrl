package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barrin, Master Wizard — Legendary Creature — Human Wizard {1}{U}{U},
// 1/1 (Urza's Saga):
//
//	"{2}, Sacrifice a permanent: Return target creature to its owner's
//	 hand."
//
// Any permanent you control pays the sacrifice, Barrin included, and so
// does the target itself (then the ability has no legal target left and
// does nothing, CR 608.2b). No {T} in the cost, so it can be activated
// the turn Barrin arrives and as often as you can pay. The sacrifice can
// be a creature, so the row declares itself a sacrifice outlet.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c6d6684-1943-4af3-98da-951604193911",
		Name:         "Barrin, Master Wizard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice a permanent: Return target creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{2}"), SacrificeAPermanent()),
			Targets: TargetCreature("target creature"),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
