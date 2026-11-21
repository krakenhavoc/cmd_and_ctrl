package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "d002391f-1dad-4966-ac36-56cc3ec015b2",
		Name:         "Ominous Cemetery",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Exile this land: Target creature's owner shuffles it into their library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{5}"), TapCost(), ExileThis()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					c, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						continue
					}
					return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: c.Owner})
				}
				return nil
			},
		}},
	})
}
