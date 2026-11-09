package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dust Animus — Creature — Spirit {1}{W}, 2/3 (Outlaws of Thunder Junction):
//
//	"Flying
//	 If you control five or more untapped lands, this creature enters with
//	 two +1/+1 counters and a lifelink counter on it.
//	 Plot {1}{W}"
//
// An entry replacement (CR 614.1c) that reads the lands as the creature
// enters: untapped lands only, and a cast that tapped them has already paid.
// The +1/+1 and lifelink counters are put on as it enters, so no state check
// sees it without them. Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "494f8c29-7626-43ed-b0b8-fb9afd4cd5d3",
		Name:            "Dust Animus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Plot("{1}{W}")},
		Replacements:    []game.ReplacementEffect{dustAnimusEntry()},
	})
}

func dustAnimusEntry() game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  "If you control five or more untapped lands, this creature enters with two +1/+1 counters and a lifelink counter on it.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID != src.InstanceID {
				return false
			}
			n := 0
			for _, c := range g.Battlefield.Cards {
				if c.Controller == src.Controller && c.IsLand() && !c.Tapped {
					n++
				}
			}
			return n >= 5
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 2)
			ev.AddCounterAtETB(game.CounterLifelink, 1)
			return nil
		},
	}
}
