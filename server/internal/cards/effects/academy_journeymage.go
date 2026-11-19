package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Academy Journeymage — Creature — Human Wizard {4}{U}, 3/2 (Foundations
// Jumpstart):
//
//	"This spell costs {1} less to cast if you control a Wizard.
//	 When this creature enters, return target creature an opponent controls
//	 to its owner's hand."
//
// Assassin's Ink's conditional self discount (CR 601.2f) over a Wizard, and
// Exclusion Mage's bounce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b400700c-1d82-4721-a166-56f88ba6ad19",
		Name:         "Academy Journeymage",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control a Wizard.", YouControlA(And(Creature(), HasSubtype("Wizard")))),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Academy Journeymage — return target creature an opponent controls to its owner's hand",
				bounceTheTargetToHand), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
