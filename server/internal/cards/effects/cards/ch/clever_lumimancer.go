package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clever Lumimancer — Creature — Human Wizard {W}, 0/1 (Strixhaven: School
// of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 this creature gets +2/+2 until end of turn."
//
// Veyran, Voice of Duality's magecraft pump at +2/+2: the shared Magecraft
// trigger (a cast or a copy, CR 207.2c) over the until-end-of-turn boost
// on the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "901eb916-f318-4a30-a43e-8e7f9eef358b",
		Name:         "Clever Lumimancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Clever Lumimancer — gets +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Clever Lumimancer — magecraft pump")),
		},
	})
}
