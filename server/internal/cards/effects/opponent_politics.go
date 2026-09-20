package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_politics.go — the shared vocabulary for cards that watch
// what an OPPONENT did, and for cards whose payoff is addressed to a
// seat that is not the source's controller.
//
// Two families live here:
//
//   - the "each turn" counters an opponent trips (their second draw,
//     their second spell, an attack at YOU with two or more
//     creatures) — Trouble in Pairs prints all three in one sentence,
//     and each of them is a read off a tally the engine already
//     keeps rather than a scan of the event log;
//   - the seat-addressed twin of a body that otherwise defaults to
//     the resolving effect's controller.
//
// Every per-turn tally the engine keeps is bumped BEFORE the event
// reaches the trigger harvester, so "their second" is `== 2` and not
// `== 1`. That is what makes these fire ONCE, on the draw or cast
// that brings the count to two, rather than on every one from the
// second onwards.

// opponentDrewTheirSecondCardThisTurn is "whenever an opponent draws
// their second card each turn": a draw by somebody other than the
// source's controller, on the event that takes that player's turn
// tally to exactly two.
//
// Caller holds g.mu.
func opponentDrewTheirSecondCardThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDrawCard || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return g.TurnTallyFor(ev.Actor).CardsDrawn == 2
}

// opponentCastTheirSecondSpellThisTurn is "whenever an opponent casts
// their second spell each turn". The cast tally is bumped before
// EventCast fires, so the second cast reads exactly two.
//
// Caller holds g.mu.
func opponentCastTheirSecondSpellThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total == 2
}

// opponentAttackedYouWithAtLeast is "whenever an opponent attacks YOU
// with N or more creatures" — Firemane Commando's shape with the
// defender pinned to the source's controller.
//
// The engine emits one EventAttack per declared attacker, so the
// ability fires on the declaration that brings the attacking player
// to N and the caller's OncePerBatch declines the rest of the batch.
// The per-turn guard behind it is belt and braces for the same thing:
// only the active player declares attackers, so one attacking seat
// per turn.
//
// An attack at your planeswalker or battle counts as an attack at
// you, because the defender is read through
// DefendingPlayerForAttackForEffect rather than off AttackingTarget
// directly.
//
// Caller holds g.mu.
func opponentAttackedYouWithAtLeast(ev game.Event, source *game.Card, g *game.Game, n int, label string) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	at := 0
	for _, id := range b13AttackingCreaturesYouControl(g, ev.Actor) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == source.Controller {
			at++
		}
	}
	if at < n {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, label)
}

// creatureDealtCombatDamageToAnOpponentOf is "whenever a creature
// deals combat damage to one of your opponents" — ANY creature, any
// controller, with the damaged player one of `controller`'s
// opponents. Edric, Spymaster of Trest and Gix, Yawgmoth Praetor
// print the same condition and hand the payoff to the damaging
// creature's controller, which the event carries in Actor.
//
// The source is looked up live: combat damage is dealt before
// state-based actions run, so an attacker that traded with its
// blocker is still on the battlefield when its damage event fires.
//
// Caller holds g.mu.
func creatureDealtCombatDamageToAnOpponentOf(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	victim := g.PlayerByIDForEffect(ev.Target)
	if victim == nil || victim.ID == controller {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.IsCreature()
}

// payLifeThenDrawFor is payLifeThenDraw addressed to a named seat:
// "that player may pay N life. If they do, they draw a card."
// payLifeThenDraw charges the resolving effect's controller, which is
// the wrong seat for a political card whose payoff belongs to an
// opponent (Gix).
//
// Paying life is a COST (CR 118.3), so "if they do" has to know the
// payment finished before the draw happens (#793), and an
// unaffordable payment degrades to doing nothing rather than erroring
// (CR 119.4) — life can move between the question and the answer.
//
// Caller holds g.mu.
func payLifeThenDrawFor(ctx *Context, player uuid.UUID, life, cards int) error {
	who := ctx.Game.PlayerByIDForEffect(player)
	if who == nil || who.Eliminated || who.Life < life {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), player, life); err != nil {
		return err
	}
	return DrawCards{Player: player, N: cards}.Apply(ctx)
}
