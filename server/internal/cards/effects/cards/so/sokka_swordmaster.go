package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "90b4d288-e950-4182-af62-516549a8712c",
		Name:            "Sokka, Swordmaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsYouControl(effects.Subtype("Ally")),
				"Equipment spells you cast cost {1} less to cast for each Ally you control.",
				effects.YourSpell(), effects.EquipmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Sokka, Swordmaster — attach up to one target Equipment you control to Sokka",
					effects.AttachTargetsToSource),
				effects.TargetPermanent("up to one target Equipment you control", effects.HasSubtype("Equipment"), effects.YouControl()).WithCount(0, 1)),
		},
	})
}
