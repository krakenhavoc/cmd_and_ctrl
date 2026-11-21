package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cactus Preserve — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add one mana of any type that a land you control could produce.
//	 {3}: Until end of turn, this land becomes an X/X green Plant
//	 creature with reach, where X is the greatest mana value among your
//	 commanders. It's still a land."
//
// The mana ability is Reflecting Pool's. X is read as the ability
// resolves (CR 608.2h), wherever the commanders are (CR 903.3); with no
// commander it is 0 and the land, a 0/0 creature, dies as a state-based
// action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8da29533-f389-4bc2-ab9b-b469f893a362",
		Name:         "Cactus Preserve",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:                    effects.ManaAbilityCost{Tap: true},
			DerivedMatch:            effects.DerivedFromOwnLands(),
			DerivedColorsOnly:       false,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any type that a land you control could produce",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Until end of turn, this land becomes an X/X green Plant creature with reach, where X is the greatest mana value among your commanders. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				x := g.GreatestCommanderManaValueForEffect(item.Controller)
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Cactus Preserve — an X/X green Plant creature with reach until end of turn",
					x, x, "Plant", []string{"G"}, "reach")
			},
		}},
	})
}
