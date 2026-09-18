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
// **The activated ability is not modelled**, and now for one reason
// rather than two. The {R/P} half is answered: #917 gave an
// activation the same announce a cast has
// (ActivateAbilityParams.PhyrexianLife, CR 602.2b, 2 life each), so
// "{1}{R/P}{R/P}" would be payable as {1} and four life the moment
// the ability exists. What is still missing is the DISCARD component
// — `AbilityCost` has no shape for "Discard two cards" (the same gap
// the Blood token declared in S21 sub-PR 4) — so the ability cannot
// be declared at all.
func init() {
	Register(Spec{
		OracleID:     "895f23a2-55b7-4cc0-8939-2efaaf097e6f",
		Name:         "Solphim, Mayhem Dominus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}{R/P}{R/P}, Discard two cards\" ability that puts an indestructible counter on Solphim can't be activated — an activation cost can't discard yet. The {R/P} half is supported."},
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
