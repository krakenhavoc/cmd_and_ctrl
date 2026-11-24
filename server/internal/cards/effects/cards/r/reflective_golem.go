package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reflective Golem — Artifact Creature — Golem {3}, 2/3 (Strixhaven: School of
// Mages, #5502):
//
//	"Whenever you cast an instant or sorcery spell that targets only this
//	 creature, you may pay {2}. If you do, copy that spell. You may choose new
//	 targets for the copy."
//
// Wandering Archaic's copy-after-pay, with the Golem's own condition: the cast
// spell is an instant or sorcery its controller casts whose every target is the
// Golem (and has at least one, CR 115.4). The trigger is read as the spell is
// cast; the {2} is the casting player's choice as the ability resolves; the copy
// is of the spell as last known if it has already resolved (CR 707.10), and the
// copy's controller may choose new targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c067dbb-934b-4ebb-a656-8541349a4b2a",
		Name:         "Reflective Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !(c.IsInstant() || c.IsSorcery()) {
					return false
				}
				spell := g.StackItemForEffect(ev.CardID)
				if spell == nil || len(spell.Targets) == 0 {
					return false
				}
				for _, t := range spell.Targets {
					if t.Kind != game.TargetCard || t.ID != source.InstanceID {
						return false
					}
				}
				return true
			}, "Reflective Golem — you may pay {2} to copy that spell", effects.MayPayTwoToCopyThatSpell("Reflective Golem — pay {2} to copy that spell?")),
		},
	})
}
