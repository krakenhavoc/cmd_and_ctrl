package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Statecraft — Enchantment {3}{U} (Mercadian Masques, #5619):
//
//	"Prevent all combat damage that would be dealt to and dealt by
//	 creatures you control."
//
// Fog Bank's standing CR 615 prevention over every creature its controller
// controls: combat damage whose source or recipient is such a creature is
// prevented, read as the damage would be dealt. Noncombat damage is not
// touched, and a player or a creature an opponent controls is hurt as
// usual by anything that is not one of your creatures.
//
// No simplification.
func init() {
	mine := func(g *game.Game, id uuid.UUID, src *game.Card) bool {
		c, ok := g.LookupCardForEffect(id)
		return ok && c.IsCreature() && c.Controller == src.Controller
	}
	effects.Register(effects.Spec{
		OracleID:     "854ef7c2-52e9-486a-aab6-7e3b3a6185ea",
		Name:         "Statecraft",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || !ev.IsCombatDamage {
					return false
				}
				return mine(g, ev.DamageSource, src) || mine(g, ev.DamageTarget, src)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Statecraft — prevent all combat damage dealt to and dealt by creatures you control",
		}},
	})
}
