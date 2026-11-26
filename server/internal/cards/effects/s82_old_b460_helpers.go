package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB460ShrineCounterTrigger is the New Phyrexia Shrine cycle's first
// ability: "At the beginning of your upkeep and whenever you cast a
// <colour> spell, put a charge counter on this artifact." It is ONE
// triggered ability with two trigger conditions (OnAny), so a restore
// names one row. The cast half is YouCast's condition, read through the
// text of the Shrine (CR 612.2).
func OldB460ShrineCounterTrigger(name, color string) game.TriggeredAbility {
	cast := YouCast(OfColor(color))
	return OnAny([]game.EventKind{game.EventBeginUpkeep, game.EventCast},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventBeginUpkeep {
				return ByYou(ev, source, lki, g)
			}
			return cast(ev, source, lki, g)
		},
		name+" — put a charge counter on this artifact",
		Do(CounterOnThis{Kind: game.CounterCharge, N: 1}))
}

// OldB460ChargeCountersOnSource is "the number of charge counters on this
// artifact" for an ability whose cost sacrificed the artifact: the count
// its last-known information holds (CR 608.2h), or the live count when it
// is somehow still there.
func OldB460ChargeCountersOnSource(g *game.Game, source uuid.UUID) int {
	if info, ok := g.LastKnownPermanentForEffect(source); ok {
		return info.Counters[game.CounterCharge]
	}
	return 0
}
