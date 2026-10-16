package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Channel Harm — Instant {5}{W}:
//
//	"Prevent all damage that would be dealt to you and permanents you
//	 control this turn by sources you don't control. If damage is prevented
//	 this way, you may have Channel Harm deal that much damage to target
//	 creature."
//
// The shield is #2026's controller test ("sources you don't control"),
// read as each source would deal damage (CR 609.7b), protecting you and
// every permanent you control. The rest is its CR 615.5 additional
// effect, once per instance of damage with that instance's total ("that
// much"), and the "you may" is asked as the damage is prevented (#2040,
// ADR 0108 amendment of 2026-10-16): the rest of the damage instruction,
// and after combat damage the state-based actions, wait for the answer.
//
// The rulings: the target is chosen as the spell is cast and carried on
// the shield (Mod.To), and only whether it is still there is checked
// afterwards; it is not a redirection — Channel Harm is the source of the
// new damage, which is not combat damage; no damage prevented (damage that
// can't be, CR 615.12) is no question asked. Two Channel Harms aimed at
// each other's creatures ask back and forth until a player declines.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     channelHarmOracle,
		Name:         "Channel Harm",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: sourceShieldSpell(PreventDamageFromSource{
			Protect: ShieldYouAndPermanentsYouControl,
			Filter:  game.DamageSourceFilter{Controller: game.SourceControllerNotYou},
			Then:    channelHarmMayDealBody,
			Label:   "Channel Harm — prevent damage from sources you don't control",
		}.DealingTo(0)),
	})
}

const channelHarmOracle = "4b3d1cfb-9f1d-4e4e-a15a-2e248ff5bdfc"

// channelHarmMayDealBody is the additional effect: "you may have Channel
// Harm deal that much damage to target creature".
var channelHarmMayDealBody = game.DelayedBody("channel-harm/may-deal-that-much", channelHarmMayDeal)

func channelHarmMayDeal(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	n := p.Amount
	if n <= 0 || len(item.Targets) != 1 {
		return nil
	}
	source, target := item.SourceCardID, item.Targets[0].ID
	name := "the creature"
	if c, ok := g.LookupCardForEffect(target); ok && c.Name != "" {
		name = c.Name
	}
	g.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      item.Controller,
		Source:       source,
		Question:     fmt.Sprintf("Channel Harm — deal %d damage to %s?", n, name),
		AcceptLabel:  fmt.Sprintf("Deal %d damage", n),
		DeclineLabel: "Don't",
		// The damage prevented was a Channel Harm's own: accepting
		// continues the loop the ruling describes, and declining ends
		// it. Bots read the hint (#2040).
		DeclineEndsLoop: channelHarmPreventedAChannelHarm(g, item),
		OnAccept: func(g *game.Game) error {
			if !g.Battlefield.Contains(target) {
				return nil
			}
			return g.DealMarkedDamageForEffect(source, nil, target, n, game.DamageMarks{})
		},
	})
	return nil
}

// channelHarmPreventedAChannelHarm reports whether the damage this
// application prevented was dealt by a Channel Harm.
func channelHarmPreventedAChannelHarm(g *game.Game, item *game.StackItem) bool {
	if item.Trigger == nil {
		return false
	}
	c, ok := g.LookupCardForEffect(item.Trigger.Event.Source)
	return ok && c.OracleID == channelHarmOracle
}
