package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swift Warden — Creature — Merfolk Warrior {1}{G}{G}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"Flash
//	 When this creature enters, target Merfolk you control gains hexproof
//	 until end of turn. (It can't be the target of spells or abilities
//	 your opponents control.)"
//
// Flash is printed data; the entry's target is chosen as the trigger goes
// on the stack and may be the Warden itself, and it gains hexproof as the
// trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0928ddda-20ea-462e-9f00-6c10ef1cfb4b",
		Name:            "Swift Warden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Swift Warden — target Merfolk you control gains hexproof until end of turn",
				effects.JmpB7FirstLegalTargetGains("Swift Warden — hexproof", "hexproof")),
				effects.TargetCreature("target Merfolk you control", effects.YouControl(), effects.Subtype("Merfolk"))),
		},
	})
}
