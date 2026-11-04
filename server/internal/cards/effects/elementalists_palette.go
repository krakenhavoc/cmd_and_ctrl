package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elementalist's Palette — Artifact {3} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Whenever you cast a spell with {X} in its mana cost, put two charge
//	 counters on this artifact.
//	 {T}: Add one mana of any color.
//	 {T}: Add {C} for each charge counter on this artifact. Spend this
//	 mana only on costs that contain {X}."
//
// The trigger reads the mana cost of the spell as cast (the face on the
// stack). The second mana ability is Everflowing Chalice's count,
// stamped with game.ManaRestrictXCost: the mana may pay the whole of a
// spell's cost when that spell's mana cost has {X}, or of an activated
// ability's cost with {X} in it, and nothing else (CR 106.6). A cast for
// an alternative cost is judged by that cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d588e25-7b4d-43c0-b45b-1b0ffb277a78",
		Name:         "Elementalist's Palette",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return strings.Contains(c.ManaCost, "{X}")
			}, "Elementalist's Palette — put two charge counters on this artifact",
				Do(CounterOnThis{Kind: game.CounterCharge, N: 2})),
		},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color",
			},
			{
				Cost:           ManaAbilityCost{Tap: true},
				ProducedAmount: everflowingChaliceProduced,
				Restrictions:   []string{game.ManaRestrictXCost},
				Label:          "{T}: Add {C} for each charge counter on this artifact (only on costs that contain {X})",
			},
		},
	})
}
