package cardse

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "6d588e25-7b4d-43c0-b45b-1b0ffb277a78",
		Name:         "Elementalist's Palette",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return strings.Contains(c.ManaCost, "{X}")
			}, "Elementalist's Palette — put two charge counters on this artifact",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 2})),
		},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color",
			},
			{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.EverflowingChaliceProduced,
				Restrictions:   []string{game.ManaRestrictXCost},
				Label:          "{T}: Add {C} for each charge counter on this artifact (only on costs that contain {X})",
			},
		},
	})
}
