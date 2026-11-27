package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prophet of Kruphix — Creature — Human Wizard {3}{G}{U}, 2/3 (Theros, S82
// old sets):
//
//	"Untap all creatures and lands you control during each other player's
//	 untap step.
//	 You may cast creature spells as though they had flash."
//
// Ivorytusk Fortress' untap-step permission over creatures and lands, and
// Yeva's creature-only flash timing with no colour filter. Both read the
// layered view, so an animated land is untapped once and a creature spell
// cast with flash is still a creature spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "372351fd-58f7-4cb8-8644-5fbc94ccb145",
		Name:         "Prophet of Kruphix",
		Completeness: effects.CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep(
				"Prophet of Kruphix — untap all creatures and lands you control",
				func(c game.Card) bool { return c.IsCreature() || c.IsLand() }),
		},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true},
				"You may cast creature spells as though they had flash."),
		},
	})
}
