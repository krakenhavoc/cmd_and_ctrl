package cardsex

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Servant — Creature — Elemental {3}{R}{R}, 4/3 (Magic 2011, #5579):
//
//	"If a red instant or sorcery spell you control would deal damage, it deals
//	 double that damage instead."
//
// A damage replacement (CR 614.1, 615) over damage whose source is a SPELL
// its controller controls (Hostility's reader, the spell's controller read
// off the damage event), an instant or sorcery, and red, all judged on the
// spell's last-known characteristics as the damage would be dealt (CR
// 608.2h). Red is a word of the Servant's text (CR 612.2). Damage from a
// permanent, from a creature spell, from a nonred spell, or from an
// opponent's spell is untouched. With a second doubler the affected player
// orders them (CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbf64afa-ace6-44b9-b47f-750c7e52cc29",
		Name:         "Fire Servant",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				if ctl, ok := g.DamageSourceSpellControllerForEffect(ev.DamageSource); !ok || ctl != src.Controller {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				return ok && (effects.HasFold(ch.Types, "Instant") || effects.HasFold(ch.Types, "Sorcery")) &&
					slices.Contains(ch.Colors, src.TextColor("R"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Fire Servant: a red instant or sorcery spell you control deals double damage",
		}},
	})
}
