package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seismic Elemental — Creature — Elemental {3}{R}{R}, 4/4 (Jumpstart,
// #5510):
//
//	"When this creature enters, creatures without flying can't block this
//	 turn."
//
// Falter's restriction on an enters trigger: it covers the creatures
// without flying as the trigger resolves (CR 611.2c), so one that is put
// onto the battlefield later this turn can still block.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61ec78da-979c-45cf-8be6-12aa24dc833a",
		Name:         "Seismic Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Seismic Elemental — creatures without flying can't block this turn",
				effects.Do(effects.RestrictUntilEOT{
					Scope:        game.ScopeCreaturesWithoutFlying,
					Restrictions: game.CantBlock,
					Label:        "Seismic Elemental — creatures without flying can't block",
				})),
		},
	})
}
