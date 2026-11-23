package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patron of the Orochi — Legendary Creature — Spirit {6}{G}{G}, 7/7
// (Betrayers of Kamigawa, #5602):
//
//	"Snake offering (You may cast this spell any time you could cast an
//	 instant by sacrificing a Snake and paying the difference in mana
//	 costs between this and the sacrificed Snake. Mana cost includes
//	 color.)
//	 {T}: Untap all Forests and all green creatures. Activate only once
//	 each turn."
//
// Snake offering is Offering (offering.go, CR 702.48). The untap covers
// every player's Forests and green creatures as the ability resolves, by
// effective subtype and colour, the Patron itself included (it is green),
// which is what lets it attack and still tap. OncePerTurnActivation is
// the printed limit (CR 602.5b), counted per object.
//
// No simplification.
func init() {
	const label = "{T}: Untap all Forests and all green creatures. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:      "e245f2f6-2b5b-41bc-88b3-5b41ca813c5e",
		Name:          "Patron of the Orochi",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Offering("Snake", effects.Subtype("Snake"))},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.TapCost(),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				forest, green := ctx.TextSubtype("Forest"), ctx.TextColor("G")
				for _, c := range effects.MatchingBattlefield(ctx, effects.Or(effects.And(effects.Land(), effects.Subtype(forest)), effects.And(effects.Creature(), effects.OfColor(green)))) {
					if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
