package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b075_helpers.go — shared bodies for OLD batch 75 (#5506).

// S82OldB075TotalCounters is the number of counters of every kind on a
// permanent ("five or more counters on it", Hero of Bretagard).
func S82OldB075TotalCounters(c *game.Card) int {
	n := 0
	for _, v := range c.Counters {
		n += v
	}
	return n
}

// S82OldB075CountersAtLeast is the AppliesTo of a static that applies to its
// own permanent while it has at least n counters of any kind on it.
func S82OldB075CountersAtLeast(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && S82OldB075TotalCounters(target) >= n
	}
}

// S82OldB075ExiledForController is one event of "a card put into exile from
// your hand, or a permanent exiled from the battlefield by a spell or
// ability you control" (Hero of Bretagard, Ranar the Ever-Watchful's
// clause): a zone move or a discard into exile.
func S82OldB075ExiledForController(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if (ev.Kind != game.EventZoneMove && ev.Kind != game.EventDiscardCard) || ev.NewZone != game.ZoneExile {
		return false
	}
	switch ev.OldZone {
	case game.ZoneHand:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Owner == controller
	case game.ZoneBattlefield:
		return game.ExiledBySpellOrAbilityOf(ev, controller)
	}
	return false
}

// S82OldB075ExiledInBatch counts the distinct cards and permanents of the
// triggering batch that S82OldB075ExiledForController accepts: "that many".
// An event with no batch stands alone.
func S82OldB075ExiledInBatch(g *game.Game, trigger game.Event, controller uuid.UUID) int {
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if trigger.Batch == 0 && ev.Seq != trigger.Seq {
			continue
		}
		if trigger.Batch != 0 && ev.Batch != trigger.Batch {
			continue
		}
		if S82OldB075ExiledForController(ev, controller, g) {
			seen[ev.CardID] = true
		}
	}
	return len(seen)
}
