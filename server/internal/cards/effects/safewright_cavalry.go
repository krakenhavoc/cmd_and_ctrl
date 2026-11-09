package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Safewright Cavalry — Creature — Elf Warrior {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"This creature can't be blocked by more than one creature.
//	 {5}: Target Elf you control gets +2/+2 until end of turn."
func init() {
	Register(Spec{
		OracleID:     "6e04a375-303e-43ba-bca1-4e959711a7f6",
		Name:         "Safewright Cavalry",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{MaxBlockers(OnSelf(), 1)},
		Activated: []ActivatedAbility{{
			Label:   "{5}: Target Elf you control gets +2/+2 until end of turn.",
			Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    ManaCost("{5}"),
			Targets: TargetCreature("target Elf you control", YouControl(), HasSubtype("Elf")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return vsPumpTheTarget(NewContext(g, item), 2, 2, "Safewright Cavalry — +2/+2 until end of turn")
			},
		}},
	})
}
