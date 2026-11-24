package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul-Scar Mage — Creature — Human Wizard {R}, 1/2:
//
//	"Prowess
//	 If a source you control would deal noncombat damage to a creature an
//	 opponent controls, put that many -1/-1 counters on that creature instead."
//
// Phytohydra's damage replacement (CR 614.1a: not a prevention, so damage that
// can't be prevented is replaced too). The damage is not dealt (no lifelink,
// no deathtouch from it, CR 614.6) and the counters go on through the counter
// window, where a doubling effect applies to them. It applies to a source you
// control as that source last existed (a spell on the stack, or a creature that
// has already left), to noncombat damage only, and to a creature an opponent
// of the Mage's controller controls; damage to a player, a planeswalker or
// your own creatures is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "009240f3-b7f8-4cbe-a3e9-974da66fb62c",
		Name:            "Soul-Scar Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: false, // CR 614.1a: a replacement, not a prevention
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || ev.IsCombatDamage || ev.DamageAmount <= 0 {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				if !ok || ch.Controller != src.Controller {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.DamageTarget)
				return ok && target.IsCreature() && target.Controller != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				n, target := ev.DamageAmount, ev.DamageTarget
				ev.Cancel()
				_, err := g.AddCounterMustSettleNowForEffect(target, game.CounterMinusOne, n)
				return err
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Soul-Scar Mage — put that many -1/-1 counters on that creature instead of dealing noncombat damage to it",
		}},
	})
}
