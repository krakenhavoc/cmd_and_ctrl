package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ominous Cemetery — Land (Doctor Who):
//
//	"{T}: Add {C}.
//	 {5}, {T}, Exile this land: Target creature's owner shuffles it into
//	 their library."
//
// The exile is part of the cost. The effect is Ice Magic's Blizzaga: the
// creature is tucked into its OWNER's library (a stolen creature goes home)
// and that owner shuffles (thenShuffle).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d002391f-1dad-4966-ac36-56cc3ec015b2",
		Name:         "Ominous Cemetery",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Exile this land: Target creature's owner shuffles it into their library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.ExileThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						continue
					}
					return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: c.Owner})
				}
				return nil
			},
		}},
	})
}
