package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cackling Slasher — Creature — Human Assassin {3}{B}, 3/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Deathtouch
//	 This creature enters with a +1/+1 counter on it if a creature died
//	 this turn."
//
// Narnam Renegade's CR 614.1c self-replacement with the condition read
// off the table-wide tally of creatures that died this turn
// (b11CreaturesDiedThisTurn), any player's and tokens included. The
// condition sits in AppliesTo, so a Slasher entering on a turn when
// nothing died offers no replacement to order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9671f4b-72d2-48d7-a497-2ea0ab72cd0b",
		Name:            "Cackling Slasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{cacklingSlasherEntry()},
	})
}

// cacklingSlasherEntry is "this creature enters with a +1/+1 counter on
// it if a creature died this turn", read as it enters.
func cacklingSlasherEntry() game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  "Enters with a +1/+1 counter on it if a creature died this turn.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				b11CreaturesDiedThisTurn(g) > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
	}
}
