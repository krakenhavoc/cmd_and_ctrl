package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arixmethes, Slumbering Isle — Legendary Creature — Kraken {2}{G}{U},
// 12/12 (EDHREC rank 2013):
//
//	"Arixmethes enters tapped with five slumber counters on it.
//	 As long as Arixmethes has a slumber counter on it, it's a land.
//	 (It's not a creature.)
//	 Whenever you cast a spell, you may remove a slumber counter from
//	 Arixmethes.
//	 {T}: Add {G}{U}.
//
// The Simic land that wakes up as a 12/12. Four pieces, all live:
//
//   - ONE entry replacement does both halves of the first line —
//     enters tapped and the five counters. Two separate
//     self-replacements would be two applicable effects on the same
//     event and would queue a CR 616 ordering prompt for a choice
//     that changes nothing.
//   - The type change is a layer 4 static on itself, gated on the
//     counter: while a slumber counter is on it its card types are
//     set to Land alone and its creature type goes with them (CR
//     205.1b), so it is not a creature — no summoning sickness on
//     the mana ability, no attacking, no dying to a wrath, and a
//     Kraken lord does nothing for it. A counter change bumps the
//     layer version, so the fifth removal wakes it at once.
//   - The cast trigger is any spell the controller casts (Arixmethes
//     itself is on the stack, not the battlefield, when it is cast,
//     so it never counts its own casting), with the "may" as a real
//     prompt; the removal is a negative AddCounter, floored at the
//     counters it has.
//   - The mana ability is a plain two-slot tap, printed colours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "caeb39ee-f0bb-4305-9e8b-b30ba0a74c78",
		Name:         "Arixmethes, Slumbering Isle",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				ev.AddCounterAtETB("slumber", 5)
				return nil
			},
			Label: "Arixmethes enters tapped with five slumber counters",
		}},
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters["slumber"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Types = []string{"Land"}
				c.Subtypes = nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, ByYou, "Arixmethes, Slumbering Isle — remove a slumber counter", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || c.Counters["slumber"] <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "slumber", N: -1}.Apply(NewContext(g, item))
			}), "Arixmethes — remove a slumber counter?"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{U}",
			Label:    "Add {G}{U}",
		}},
	})
}
