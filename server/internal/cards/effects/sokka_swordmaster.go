package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sokka, Swordmaster — Legendary Creature — Human Warrior Ally {2}{W},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Vigilance
//	 Equipment spells you cast cost {1} less to cast for each Ally you
//	 control.
//	 At the beginning of combat on your turn, attach up to one target
//	 Equipment you control to Sokka."
//
// The reduction counts every Ally you control as the cost is determined
// (CR 601.2f), Sokka himself included, as the ruling has it; it reduces
// only generic mana. The combat trigger is Captain America, First
// Avenger's (AttachTargetsToSource).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "90b4d288-e950-4182-af62-516549a8712c",
		Name:            "Sokka, Swordmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		CostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(Subtype("Ally")),
				"Equipment spells you cast cost {1} less to cast for each Ally you control.",
				YourSpell(), equipmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Sokka, Swordmaster — attach up to one target Equipment you control to Sokka",
					AttachTargetsToSource),
				TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1)),
		},
	})
}
