package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Seedcore — Land — Sphere (EDHREC rank 5539):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast
//	 Phyrexian creature spells.
//	 Corrupted — {T}: Target 1/1 creature gets +2/+1 until end of turn.
//	 Activate only if an opponent has three or more poison counters."
//
// Phyrexia: All Will Be One's toxic-deck utility land. The coloured
// mana carries the cast / creature / Phyrexian restriction tags, and
// the auto-tapper leaves it alone. Corrupted is the pump's activation
// condition (CR 602.1b, #743): some one opponent has three or more
// poison counters, which are public. "Target 1/1 creature" is the
// creature's current power and toughness at announce, re-checked on
// resolution like any target, so a creature pumped in response is no
// longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "249fdd3e-376c-4ec2-a612-4353e0e61ee2",
		Name:         "The Seedcore",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color (Phyrexian creature spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictType("Creature"),
					ManaRestrictSubtype("Phyrexian"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "Corrupted — {T}: Target 1/1 creature gets +2/+1 until end of turn. Activate only if an opponent has three or more poison counters.",
			Cost:  TapCost(),
			Targets: TargetCreature("target 1/1 creature", func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.CurrentPower() == 1 && c.CurrentToughness() == 1
			}),
			Condition: seedcoreCorrupted,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: 2, Toughness: 1, Label: "The Seedcore — +2/+1"}.Apply(ctx)
			},
		}},
	})
}

// seedcoreCorrupted is corrupted: an opponent has three or more poison
// counters.
func seedcoreCorrupted(g *game.Game, controller, _ uuid.UUID) bool {
	return eachOpponent(g, controller, func(opp uuid.UUID) bool {
		p := g.PlayerByIDForEffect(opp)
		return p != nil && p.Poison >= 3
	})
}
