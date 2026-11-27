package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rappelling Scouts — Creature — Human Rebel Scout {2}{W}{W}, 1/4 (Mercadian
// Masques, S82 OLD b682, #5619):
//
//	"Flying
//	 {2}{W}: This creature gains protection from the color of your choice
//	 until end of turn."
//
// Flying is a printed keyword. The colour is chosen as the ability resolves
// (Alseid of Life's Bounty's choice, CR 608.2d), for the Scouts themselves,
// and the protection lasts until end of turn. A Scout gone by then gets
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c71eb9b5-2017-4e6c-8140-f9c8af65893f",
		Name:            "Rappelling Scouts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: This creature gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{2}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				effects.ChooseProtectionColorFor(effects.NewContext(g, item), item.Controller, item.SourceCardID, "Rappelling Scouts", "Rappelling Scouts — choose a color")
				return nil
			},
		}},
	})
}
