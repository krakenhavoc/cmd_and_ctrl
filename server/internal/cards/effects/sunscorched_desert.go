package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunscorched Desert — Land — Desert (EDHREC rank 3293):
//
//	"When this land enters, it deals 1 damage to target player or
//	 planeswalker.
//	 {T}: Add {C}."
//
// The Amonkhet Desert that pings on arrival. The entry trigger is
// b06SelfETB with Boros Charm's "target player or planeswalker"
// clause, picked when the trigger goes on the stack and re-checked
// at resolution (CR 608.2b); the damage is dealt by the Desert, a
// colorless source. The mana ability is a plain colorless tap, and
// the Desert subtype is on the type line for Scavenger Grounds and
// Hour of Promise to read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "256b8c23-589e-429d-9e6e-433d55079eb4",
		Name:         "Sunscorched Desert",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   targetPlayerOrPlaneswalker("target player or planeswalker"),
			Key:       "Sunscorched Desert — 1 damage to target player or planeswalker",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
				}
				return nil
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
