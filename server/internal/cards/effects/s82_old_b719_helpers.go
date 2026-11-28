package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b719_helpers.go — shared pieces of the S82 OLD batch 719 cards
// (Urza's Legacy, #5623).

// OldB719CantBeBlockedWhileDefenderControls is "This creature can't be blocked
// as long as defending player controls <a permanent that match accepts>"
// (Bouncing Beebles). The defending player is the controller of any creature
// that could block it (CR 506.3: only that player's creatures may), so the
// pair rule reads the blocker's controller's battlefield, as the block is
// declared (CR 509.1b).
func OldB719CantBeBlockedWhileDefenderControls(match CardPredicate) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBeBlocked,
		Pair: func(g *game.Game, attacker, blocker, source *game.Card) bool {
			if source == nil || attacker.InstanceID != source.InstanceID {
				return false
			}
			defender := blocker.Controller
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == defender && match(g, defender, c) {
					return true
				}
			}
			return false
		},
	}
}

// OldB719SacrificeTheEnchantedCreatureUnlessPay is Slow Motion's
// "At the beginning of the upkeep of enchanted creature's controller, that
// player sacrifices that creature unless they pay {2}": the upkeep's player is
// asked, and declining (or a payment they cannot fund) sacrifices the creature
// the Aura enchanted as the trigger resolved. The creature rides the prompt as
// plain data, so a table waiting on the payer is a restore point.
func OldB719SacrificeTheEnchantedCreatureUnlessPay(name, cost string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item == nil || item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		aura, ok := ctx.SourcePermanent()
		if !ok || aura.AttachedTo.Kind != game.TargetCard {
			return nil
		}
		host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
		if !ok || !OnBattlefield(g, host.InstanceID) {
			return nil
		}
		return UpkeepPayUnless{
			Chooser:  item.Trigger.Event.Actor,
			Cost:     cost,
			Question: name + " — pay " + cost + ", or sacrifice " + host.Name + "?",
			Else:     oldB719SacrificeTheCarried,
			Carry:    []uuid.UUID{host.InstanceID},
		}.Apply(ctx)
	}
}

var oldB719SacrificeTheCarried = PayUnlessElse("pay-unless/old-b719-sacrifice-the-enchanted-creature",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) {
			return nil
		}
		return SacrificePermanent{Target: r.Carry[0]}.Apply(ctx)
	})
