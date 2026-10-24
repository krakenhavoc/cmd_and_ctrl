package effects

import (
	"fmt"

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

// targetPlayerGetsRadCounters is "target player gets n rad counters"
// (Nightkin Ambusher): the player the item targets, if that target is
// still legal as it resolves (CR 608.2b).
func targetPlayerGetsRadCounters(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		target := targetedPlayerOf(NewContext(g, item))
		if target == uuid.Nil {
			return nil
		}
		return playerGetsRadCounters(g, item.Controller, target, n)
	}
}

// DefendingPlayerHasARadCounter is "as long as defending player has a
// rad counter", read on the attacker being blocked (Nightkin Ambusher):
// the defending player is the one behind the attacker's attack target
// (CR 508.5). An attacker that is not attacking has no defending
// player, and the clause is false.
func DefendingPlayerHasARadCounter(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
	defender := g.DefendingPlayerForAttackForEffect(attacker.AttackingTarget)
	return defender != uuid.Nil && g.RadCountersOf(defender) > 0
}

// radCountersAmongPlayers is "the total number of rad counters among
// players": every rad counter every player still in the game has.
func radCountersAmongPlayers(g *game.Game) int {
	total := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		total += p.Counters[game.CounterRad]
	}
	return total
}

// mayEnterTappedToGetRadCounters is "You may have this land enter
// tapped. If you do, you get n rad counters" (Mariposa Military Base):
// an optional enters-tapped replacement whose "yes" is the whole
// payment. The yes/no is put to the player performing the entry, as a
// shockland's is (EnteringPermanentChooser). On "yes" the land is
// tapped as it enters and the player gets the counters, through the
// CR 614 player-counter window; on "no", or in a window that cannot
// ask, nothing happens and the land enters untapped.
func mayEnterTappedToGetRadCounters(name string, n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:        fmt.Sprintf("You may have this land enter tapped. If you do, you get %d rad counters.", n),
		EntryOnly:       true,
		Optional:        true,
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  fmt.Sprintf("%s — enter tapped and get %d rad counters?", name, n),
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			ev.EntersTapped = true
			player := EnteringPermanentChooser(ev, g, src)
			if player == uuid.Nil {
				return nil
			}
			return playerGetsRadCounters(g, player, player, n)
		},
	}
}

// radCountersYouHave is the amount for "costs {1} less to activate for
// each rad counter you have": the activator's rad counters at the time
// the cost is determined (CR 602.2b, 601.2f). It never goes below zero.
func radCountersYouHave(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return q.Game.RadCountersOf(q.Controller)
}
