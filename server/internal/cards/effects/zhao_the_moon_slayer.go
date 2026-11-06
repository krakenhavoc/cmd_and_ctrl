package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhao, the Moon Slayer — Legendary Creature — Human Soldier {1}{R}, 2/2:
//
//	"Menace
//	 Nonbasic lands enter tapped.
//	 {7}: Put a conqueror counter on Zhao.
//	 As long as Zhao has a conqueror counter on him, nonbasic lands are
//	 Mountains. (They lose all other land types and abilities and have
//	 "{T}: Add {R}.")"
//
// The second line is Archon of Emeria's enters-tapped replacement
// (CR 614.12) for every nonbasic land, whoever controls it. It applies
// on top of any "enters tapped unless" the land prints, so such a land
// enters tapped even when its condition holds (the ruling).
//
// The last line is Blood Moon's layer-4 land-type setter (CR 305.7)
// gated on the counter, read live (CR 611.3a): while Zhao has a
// conqueror counter, a nonbasic land loses its other land types and
// abilities, so an enters trigger it prints never triggers (the
// ruling). Names and supertypes are untouched.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "94c64d41-d756-42b6-b0e7-0a351d64a68c",
		Name:            "Zhao, the Moon Slayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
					return false
				}
				entering, ok := g.LookupCardForEffect(ev.CardID)
				return ok && b10NonbasicLand()(g, uuid.Nil, entering)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Zhao, the Moon Slayer: nonbasic lands enter tapped",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{7}: Put a conqueror counter on Zhao.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{7}"),
			Effect:  Do(CounterOnThis{Kind: zhaoConquerorCounter, N: 1}),
		}},
		Static: []game.StaticAbility{
			Labelled(SetsBasicLandType(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return source.Counters[zhaoConquerorCounter] > 0 && nonbasicLand(target, g, source)
			}, nil, []string{"Mountain"}), "As long as Zhao has a conqueror counter on him, nonbasic lands are Mountains."),
		},
	})
}

const zhaoConquerorCounter = "conqueror"
