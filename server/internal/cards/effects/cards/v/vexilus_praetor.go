package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vexilus Praetor — Creature — Custodes Warrior {3}{W}, 3/4 (Warhammer 40,000
// Commander, S82 GAP batch 196, #5948):
//
//	"Flash
//	 Vigilance
//	 Aegis of the Emperor — Commanders you control have protection from
//	 everything."
//
// A layer 6 grant of "protection from everything" (CR 702.16j) to the commanders
// its controller controls on the battlefield: such a commander can't be blocked,
// targeted, dealt damage or enchanted or equipped by anything, and it deals and
// receives no damage from sources (the engine's protection reads the keyword as
// it does any other quality). It doesn't reach a commander in the command zone or
// a card that is not a commander.
//
// No simplification.
func init() {
	grant := effects.B16GrantKeywords(effects.GapB196CommandersYouControl, "protection from everything")
	grant.Label = "Commanders you control have protection from everything."
	effects.Register(effects.Spec{
		OracleID:        "e3a65df1-ba3a-4635-a321-8cab9a1f0740",
		Name:            "Vexilus Praetor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance"},
		Static:          []game.StaticAbility{grant},
	})
}
