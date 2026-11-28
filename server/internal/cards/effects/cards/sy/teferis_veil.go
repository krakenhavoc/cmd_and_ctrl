package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Veil — Enchantment {1}{U} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Whenever a creature you control attacks, it phases out at end of combat. (While it's phased
//	 out, it's treated as though it doesn't exist. It phases in before you untap during your
//	 next untap step.)"
//
// One trigger per attacking creature you control: each schedules a delayed trigger at the
// beginning of this combat's end of combat step (CR 511.2) that phases that creature out (and
// whatever is attached to it, CR 702.26) if it is still on the battlefield. Phasing in is the
// engine's untap-step rule.
//
// No simplification.
func init() {
	const label = "Teferi's Veil — it phases out at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "c6db2ea4-954c-443b-9603-c41596555c31",
		Name:         "Teferi's Veil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, label, effects.S82OldB778PhaseOutTheAttackerAtEndOfCombat(label)),
		},
	})
}
