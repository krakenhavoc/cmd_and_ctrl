package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rad_counters.go — the shared vocabulary of the rad-counter cards
// (#2042). Append-only and mechanic-named, so the clone gate never sees
// the same body twice.
//
// The counters themselves do their work in the engine: CR 728.1's
// inherent trigger (game/rad_counters.go) mills at the beginning of a
// player's precombat main phase and costs them life for each nonland
// card. What a card does is GIVE them, and every gift goes through the
// CR 614 player-counter window with the giver named (CR 120.3b's "you
// give"), as Ichor Rats' poison does.

// playerGetsRadCounters is "<player> gets n rad counters", given by
// `giver` (the item's controller). A player who has left the game gets
// nothing (CR 800.4a).
func playerGetsRadCounters(g *game.Game, giver, player uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
		return nil
	}
	return g.AddPlayerCounterByForEffect(giver, player, game.CounterRad, n)
}

// eachPlayerGetsRadCounters is "each player gets n rad counters", in
// seat order.
func eachPlayerGetsRadCounters(g *game.Game, giver uuid.UUID, n int) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if err := playerGetsRadCounters(g, giver, p.ID, n); err != nil {
			return err
		}
	}
	return nil
}

// eachOpponentGetsRadCounters is "each opponent gets n rad counters":
// every player still in the game other than `giver`.
func eachOpponentGetsRadCounters(g *game.Game, giver uuid.UUID, n int) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == giver {
			continue
		}
		if err := playerGetsRadCounters(g, giver, p.ID, n); err != nil {
			return err
		}
	}
	return nil
}

// damagedPlayerGetsRadCounters is "Whenever this creature deals combat
// damage to a player, they get n rad counters" (Glowing One): the
// damaged player is the triggering event's target.
func damagedPlayerGetsRadCounters(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		return playerGetsRadCounters(g, item.Controller, item.Trigger.Event.Target, n)
	}
}

// ANonlandCardWasMilled is "whenever a player mills a nonland card": the
// engine emits one EventMill per card that reached a graveyard, and the
// card is read where it landed (CR 701.17c).
func ANonlandCardWasMilled(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventMill {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsLand()
}
