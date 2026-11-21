package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// route_batch_restore_test.go — #4253 slice 9: a route batch (an exile, a
// mill, a discard, manifest dread's two moves, a discard that became an
// entry) carries the rest of the batch as data, so a table waiting on a CR
// 616 prompt in the middle of it is a restore point. Each test pauses the
// batch on an ordering prompt over its SECOND card, takes a STRICT restore
// (the boot path), answers on the restored table, and checks the first
// card stayed where it went, the second moved after the answer, and the
// card's own "then" ran off the disk.

// graveyardOrderProbeOracles are two test cards, each "if a card named
// Needle would be put into a graveyard, it is put into a graveyard
// instead": two declared effects, so a Needle's move to a graveyard (a
// mill, a discard, a plain move) asks for their order (CR 616.1).
var GraveyardOrderProbeOracles = []string{"test-4253-graveyard-order-a", "test-4253-graveyard-order-b"}

// entryOrderProbeOracles are the same over an Obstinate Baloth entering
// the battlefield.
var EntryOrderProbeOracles = []string{"test-4253-entry-order-a", "test-4253-entry-order-b"}

func init() {
	for i, oracle := range GraveyardOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Graveyard Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if (ev.Kind != game.RepEventMove && ev.Kind != game.RepEventDiscard) || ev.NewZone != game.ZoneGraveyard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Graveyard Order Probe " + string(rune('A'+i)),
			}},
		})
	}
	for i, oracle := range EntryOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Entry Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Obstinate Baloth"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Entry Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

func SeedOrderProbes(g *game.Game, who uuid.UUID, oracles []string, name string) {
	for i, oracle := range oracles {
		SeedReplacementPermanent(g, oracle, name+" "+string(rune('A'+i)), who)
	}
}

func SeatByID(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// A random discard with a registered "then", told the cards discarded this
// way: the keyed door's tail runs off the disk after a restore mid-batch.
var TestDiscardedGainLife = effects.CardTail("test/4253-discarded-gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	return ctx.Game.ChangePlayerLifeForEffect(uuid.Nil, ctx.Controller(), len(o.IDs))
})
