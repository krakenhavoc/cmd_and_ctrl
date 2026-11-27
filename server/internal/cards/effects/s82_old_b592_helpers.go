package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b592_helpers.go — Saviors of Kamigawa, batch 592 (#5601): the
// hand-size counts the Kami's legends read, Sokenzan Renegade's "a player has
// more cards in hand than each other player", Sekki's prevention follow-up and
// the pair of block/become-blocked readings Shape Stealer shares.

// OldB592MostCardsInAnOpponentsHand is "the number of cards in the hand of the
// opponent with the most cards in hand" (Adamaro): the largest hand among the
// source controller's opponents still in the game, zero with none left. Read
// live inside the layer pass, so it walks the seats directly.
func OldB592MostCardsInAnOpponentsHand(g *game.Game, you uuid.UUID) int {
	best := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you || p.Hand == nil {
			continue
		}
		best = max(best, p.Hand.Size())
	}
	return best
}

// OldB592HandLeader is "a player has more cards in hand than each other
// player" (Sokenzan Renegade): the one player, among those still in the game,
// whose hand is strictly larger than every other's. Two players tied at the top
// mean nobody has more than each other player (uuid.Nil, false).
func OldB592HandLeader(g *game.Game) (uuid.UUID, bool) {
	var best uuid.UUID
	bestSize, tied, found := 0, false, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Hand == nil {
			continue
		}
		size := p.Hand.Size()
		switch {
		case !found || size > bestSize:
			best, bestSize, tied, found = p.ID, size, false, true
		case size == bestSize:
			tied = true
		}
	}
	if !found || tied {
		return uuid.Nil, false
	}
	return best, true
}

// OldB592HandLeaderExists is the trigger-time half of Sokenzan Renegade's
// intervening "if a player has more cards in hand than each other player".
func OldB592HandLeaderExists(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := OldB592HandLeader(g)
	return ok
}

// OldB592HandLeaderCondition is the same condition as the trigger resolves
// (CR 603.4).
func OldB592HandLeaderCondition() Condition {
	return CondBoard{
		Name: "a player has more cards in hand than each other player",
		Test: func(ctx *Context) bool {
			_, ok := OldB592HandLeader(ctx.Game)
			return ok
		},
	}
}

// OldB592PairOtherCreature is, for a block event (one per blocker and the
// attacker it blocks, CR 509.3b and 509.3d), the creature of the pair that is
// not `self`: the attacker when `self` blocks, the blocker when `self` is
// blocked. It reports false when the event is not a block of `self`.
func OldB592PairOtherCreature(ev game.Event, self uuid.UUID) (uuid.UUID, bool) {
	if ev.Kind != game.EventBlock {
		return uuid.Nil, false
	}
	switch self {
	case ev.CardID:
		return ev.Target, true
	case ev.Target:
		return ev.CardID, true
	}
	return uuid.Nil, false
}

// OldB592SekkiSpiritsBody is the follow-up of Sekki's prevention: "remove that
// many +1/+1 counters from Sekki and create that many 1/1 colorless Spirit
// creature tokens". The counters come off up to what Sekki has; the tokens are
// made for all the damage that was prevented (the damage is Sekki's own, the
// count printed as "that many"). Sekki being gone by then makes neither. An
// on-disk key.
var OldB592SekkiSpiritsBody = game.DelayedBody("prevention/s82-old-b592-sekki-counters-for-spirits", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := FollowUpThis(g, item)
	if !ok {
		return nil
	}
	n := ThatDamage(item)
	if n <= 0 {
		return nil
	}
	if remove := min(n, info.Counters[game.CounterPlusOne]); remove > 0 {
		if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -remove); err != nil {
			return err
		}
	}
	return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: n}.Apply(NewContext(g, item))
})
