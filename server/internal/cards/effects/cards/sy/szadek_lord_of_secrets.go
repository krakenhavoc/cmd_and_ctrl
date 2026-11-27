package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Szadek, Lord of Secrets — Legendary Creature — Vampire {3}{U}{U}{B}{B}, 5/5
// (Commander 2011, S82 older sets batch 454, #5574):
//
//	"Flying
//	 If Szadek would deal combat damage to a player, instead put that many
//	 +1/+1 counters on Szadek and that player mills that many cards."
//
// A replacement effect on combat damage Szadek deals to a player (Soul-Scar
// Mage's shape, CR 614.1a): the damage is not dealt, so nothing that needs it
// dealt (lifelink, "deals combat damage" triggers) happens. The amount is what
// the event holds after any earlier prevention or doubling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c40c823c-5954-457a-81a6-02683da57f7f",
		Name:            "Szadek, Lord of Secrets",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: false, // CR 614.1a: a replacement, not a prevention
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageAmount <= 0 ||
					ev.DamageSource != src.InstanceID {
					return false
				}
				return g.PlayerByIDForEffect(ev.DamageTarget) != nil
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				n, player := ev.DamageAmount, ev.DamageTarget
				ev.Cancel()
				if _, err := g.AddCounterMustSettleNowForEffect(src.InstanceID, game.CounterPlusOne, n); err != nil {
					return err
				}
				return g.MillNForEffect(player, n)
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Szadek, Lord of Secrets — put that many +1/+1 counters on Szadek and that player mills that many cards instead of dealing combat damage",
		}},
	})
}
