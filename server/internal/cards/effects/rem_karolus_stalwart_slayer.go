package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rem Karolus, Stalwart Slayer — Legendary Creature — Human Knight
// {1}{R}{W}, 2/3 (Midnight Hunt Commander):
//
//	"Flying, haste
//	 If a spell would deal damage to you or another permanent you control,
//	 prevent that damage.
//	 If a spell would deal damage to an opponent or a permanent an
//	 opponent controls, it deals that much damage plus 1 instead."
//
// Two replacement rows read the damage event's source: it must be a spell
// (a card in the stack zone, s82MidB17SourceIsSpell), so combat damage
// and a permanent's ability are untouched. The first is a prevention
// (CR 615.1a, so "can't be prevented" beats it): it covers the controller
// and every permanent they control except Rem Karolus himself. The second
// is Torbran's shape with a +1 and a spell gate. The two never overlap: a
// recipient is either yours or an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21971c8b-ff9b-40f8-8c9a-93d26c2615b7",
		Name:            "Rem Karolus, Stalwart Slayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Replacements: []game.ReplacementEffect{
			{
				Watches:    []game.EventKind{game.EventDealDamage},
				Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil {
						return false
					}
					if !s82MidB17SourceIsSpell(g, ev.DamageSource) {
						return false
					}
					if ev.DamageTarget == src.Controller {
						return true
					}
					c, ok := g.LookupCardForEffect(ev.DamageTarget)
					return ok && g.Battlefield.Contains(ev.DamageTarget) &&
						c.Controller == src.Controller && c.InstanceID != src.InstanceID
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.Cancel()
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Rem Karolus — prevent spell damage to you and your other permanents",
			},
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
						return false
					}
					if !s82MidB17SourceIsSpell(g, ev.DamageSource) {
						return false
					}
					return damageHitsAnOpponentOf(ev, g, src.Controller)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.DamageAmount++
					return nil
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Rem Karolus — spells deal 1 more damage to opponents and their permanents",
			},
		},
	})
}
