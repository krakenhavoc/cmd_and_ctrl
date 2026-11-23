package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aetherwind Basker — Creature — Lizard {4}{G}{G}{G}, 7/7:
//
//	"Trample
//	 Whenever this creature enters or attacks, you get {E} (an energy
//	 counter) for each creature you control.
//	 Pay {E}: This creature gets +1/+1 until end of turn."
//
// ADR 0129 PR 1 (#1995). The creatures are counted as the trigger
// resolves, the Basker included while it is on the battlefield. The
// amount is counted, so it is not declared as a purpose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d9ec9a4f-0644-47c5-bc09-1201d5251bde",
		Name:            "Aetherwind Basker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Aetherwind Basker — you get {E} for each creature you control",
				effects.EbYouGetEnergyPerCreatureYouControl),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.PayEnergy(1),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Aetherwind Basker — +1/+1 until end of turn"),
		}},
	})
}
