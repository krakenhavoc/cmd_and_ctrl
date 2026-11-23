package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mass_route_restore_test.go — #4253 slice 10a: the mass verbs, the single
// sacrifice and bounce, the tuck and destroy callers, the exile whose "then"
// grants a cast permission or records an "until" return, an exchange's exit
// and a mill's amount window are restore points. Each test pauses on a CR
// 616 ordering prompt, takes a STRICT restore (the boot path), answers on
// the restored table, and checks the rest of the card ran off the disk.

// libraryHandOrderProbeOracles are two test cards, each "if a card named
// Needle would be put into a library or a hand, it is put there instead":
// two declared effects, so a Needle's tuck or bounce asks for their order
// (CR 616.1).
var LibraryHandOrderProbeOracles = []string{"test-4253-10a-library-hand-order-a", "test-4253-10a-library-hand-order-b"}

// millPlusOneProbeOracles are two test cards, each "if a player would mill
// one or more cards, they mill that many plus one instead": two declared
// mill-amount replacements, so a mill's amount window asks the milling
// player for their order (CR 616.1).
var MillPlusOneProbeOracles = []string{"test-4253-10a-mill-plus-one-a", "test-4253-10a-mill-plus-one-b"}

func init() {
	for i, oracle := range MillPlusOneProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Mill Plus One Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventMill},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMill && ev.MillCount > 0
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.MillCount++
					return nil
				},
				Label: "Mill Plus One Probe " + string(rune('A'+i)),
			}},
		})
	}
	for i, oracle := range LibraryHandOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Library Hand Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventMove || (ev.NewZone != game.ZoneLibrary && ev.NewZone != game.ZoneHand) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Name == "Needle"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Library Hand Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

// tableSeats is the active seat and the one after it.
func TableSeats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}
