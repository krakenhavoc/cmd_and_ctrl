package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naban, Dean of Iteration — Legendary Creature — Human Wizard {1}{U}, 2/1
// (Foundations Jumpstart):
//
//	"If a Wizard you control entering causes a triggered ability of a
//	 permanent you control to trigger, that ability triggers an additional
//	 time."
//
// Traveling Chocobo's doubler (DoublesEntering) over a Wizard under your
// control. Naban is a Wizard, so his own entry and the Wizards after him
// count. The Wizard word is read from the source's text, as any creature
// type.
//
// No simplification.
func init() {
	doubler := effects.DoublesEntering(effects.And(effects.YouControl(), effects.OfCreatureType("Wizard")))
	doubler.Label = "Naban, Dean of Iteration"
	doubler.RowLabel = "Abilities that trigger when a Wizard you control enters trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "cffb9b7f-d52c-43e0-a4a6-b06dea83eb53",
		Name:            "Naban, Dean of Iteration",
		Completeness:    effects.CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
