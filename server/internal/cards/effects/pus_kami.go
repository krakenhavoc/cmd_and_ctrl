package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pus Kami — Creature — Spirit {5}{B}{B}, 3/3 (Betrayers of Kamigawa,
// #5602):
//
//	"{B}, Sacrifice this creature: Destroy target nonblack creature.
//	 Soulshift 6 (When this creature dies, you may return target Spirit
//	 card with mana value 6 or less from your graveyard to your hand.)"
//
// Sacrificing it to its own ability is a death, so soulshift (the shared
// Soulshift trigger, CR 702.46a) fires too and can return a Spirit card
// with mana value 6 or less, though not Pus Kami itself (mana value 7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b2471d4-2821-4220-becc-94bebfdc456f",
		Name:         "Pus Kami",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Soulshift(6, "Pus Kami")},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice this creature: Destroy target nonblack creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{B}"), SacrificeThis()),
			Targets: TargetCreature("target nonblack creature", NonBlack()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
