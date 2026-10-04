package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch's Clinic — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Target commander gains lifelink until end of turn."
//
// "Target commander" is ANY player's commander on the battlefield
// (Card.IsCommander, CR 903.3), not only the controller's. A commander
// in the command zone or on the stack is not a legal target here, which
// loses nothing: lifelink granted until end of turn does not follow a
// card to the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05899372-9784-4bdb-9c28-504c71fed906",
		Name:         "Witch's Clinic",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Target commander gains lifelink until end of turn",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPermanent("target commander",
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCommander }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (GrantKeywordUntilEOT{
						Target:   target.ID,
						Keywords: []string{"lifelink"},
						Label:    "Witch's Clinic — lifelink",
					}).Apply(ctx)
				}
				return nil
			},
		}},
	})
}
