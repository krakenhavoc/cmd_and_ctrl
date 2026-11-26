package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arashin Foremost — Creature — Human Warrior {1}{W}{W}, 2/2 (Dragons of Tarkir, #5551):
//
//	"Double strike
//	 Whenever this creature enters or attacks, another target Warrior creature you control gains double strike until end of turn."
//
// One ability with two conditions (Sun Titan's WhenThisEntersOrAttacks); the
// target is another Warrior, so the Foremost is never its own target.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "581c2fb6-fcce-42c1-b198-1f8d62e862b1",
		Name:            "Arashin Foremost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Arashin Foremost — another target Warrior creature you control gains double strike until end of turn",
					effects.EbTargetCreatureGainsUntilEOT("double strike", "Arashin Foremost — double strike")),
				effects.Another(effects.TargetCreature("another target Warrior creature you control", effects.YouControl(), effects.HasSubtype("Warrior")))),
		},
	})
}
