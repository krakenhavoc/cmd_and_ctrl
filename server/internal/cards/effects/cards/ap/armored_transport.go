package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armored Transport — Artifact Creature — Construct {3}, 2/1 (Gatecrash,
// #5565):
//
//	"Prevent all combat damage that would be dealt to this creature by
//	 creatures blocking it."
//
// A standing CR 615 prevention (Marble Priest's shape) narrowed to combat
// damage whose source is a creature that is blocking this one when it is
// dealt; combat damage from creatures that are not blocking it
// is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0caae56e-5995-4f1b-b735-80580a372707",
		Name:         "Armored Transport",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageTarget != src.InstanceID {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.DamageSource)
				return ok && dealer.IsCreature() && dealer.IsBlockingAttacker(src.InstanceID)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Armored Transport — prevent combat damage dealt to it by creatures blocking it",
		}},
	})
}
