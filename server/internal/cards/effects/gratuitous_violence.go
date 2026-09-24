package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gratuitous Violence — Enchantment {2}{R}{R}{R} (EDHREC rank 1513):
//
//	"If a creature you control would deal damage to a permanent or
//	 player, it deals double that damage instead."
//
// Angrath's Marauders' replacement narrowed to CREATURE sources: a
// Bolt is not doubled, an attacker is. The source's type and
// controller are read from the damage event's last-known
// characteristics (#1430, CR 608.2h) — the same value protection and
// the catalog's colour-testing doublers read — so a creature that
// traded is still on the battlefield when its damage is replaced,
// and a creature that dealt "when this dies" damage after changing
// type or control is still judged by what it was when it died, not
// by its graveyard card. With a second doubler the affected player
// orders them, and x4 is x4 either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c340a39-4ee0-4ba1-bb66-6674f8020fda",
		Name:         "Gratuitous Violence",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				return ok && hasFold(ch.Types, "Creature") && ch.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gratuitous Violence: double the damage",
		}},
	})
}
