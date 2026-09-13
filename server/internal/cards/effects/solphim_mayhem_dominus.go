package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solphim, Mayhem Dominus — 5/4 Legendary Creature — Phyrexian
// Horror for {2}{R}{R}:
//
//	"If a source you control would deal noncombat damage to an
//	 opponent or a permanent an opponent controls, it deals double
//	 that damage to that player or permanent instead.
//	 {1}{R/P}{R/P}, Discard two cards: Put an indestructible counter
//	 on Solphim."
//
// Angrath's Marauders' doubler with two restrictions the Marauders
// lack: noncombat only, and opponents only. Both are expressible —
// `ReplacementEvent.IsCombatDamage` and a controller check on the
// damage target — so this is the narrower, more accurate version of
// a shape the catalog already proved.
//
// Doubling every Fireweaver ping and every Magmakin trigger is what
// this deck wants it for; it deliberately does nothing for the
// Pirates' combat damage.
//
// **The activated ability is not modelled.** `AbilityCost` has no
// discard component (the same gap the Blood token declared in S21
// sub-PR 4), and the {R/P} Phyrexian mana would need a pay-2-life
// alternative the cost parser records but doesn't offer. Both are
// one piece of work; this card is now the third asking for it.
func init() {
	Register(Spec{
		OracleID:     "895f23a2-55b7-4cc0-8939-2efaaf097e6f",
		Name:         "Solphim, Mayhem Dominus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}{R/P}{R/P}, Discard two cards\" ability that puts an indestructible counter on Solphim can't be activated."},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage {
					return false
				}
				if !damageSourceControlledBy(ev, g, src.Controller) {
					return false
				}
				return damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
		}},
	})
}

// damageHitsAnOpponentOf reports whether the damage target is a
// player other than `you`, or a permanent someone else controls.
// The target ID is a player or a card and the kind isn't recorded,
// so this resolves it as a player first and falls back to a card
// lookup — the same disambiguation the damage emitter uses.
func damageHitsAnOpponentOf(ev *game.ReplacementEvent, g *game.Game, you uuid.UUID) bool {
	if ev.DamageTarget == uuid.Nil {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.DamageTarget); p != nil {
		return p.ID != you
	}
	c, ok := g.LookupCardForEffect(ev.DamageTarget)
	return ok && c.Controller != you
}
