package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "05899372-9784-4bdb-9c28-504c71fed906",
		Name:         "Witch's Clinic",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target commander gains lifelink until end of turn",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target commander",
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCommander }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, target := range ctx.LegalTargets() {
					if target.Kind != game.TargetCard {
						continue
					}
					return (effects.GrantKeywordUntilEOT{
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
