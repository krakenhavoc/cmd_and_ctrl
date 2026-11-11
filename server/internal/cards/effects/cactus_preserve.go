package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8da29533-f389-4bc2-ab9b-b469f893a362",
		Name:         "Cactus Preserve",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			DerivedMatch:            DerivedFromOwnLands(),
			DerivedColorsOnly:       false,
			DerivesFromOtherSources: true,
			Label:                   "Add one mana of any type that a land you control could produce",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Until end of turn, this land becomes an X/X green Plant creature with reach, where X is the greatest mana value among your commanders. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				x := g.GreatestCommanderManaValueForEffect(item.Controller)
				return cmdr2B13Animate(NewContext(g, item), item.SourceCardID,
					"Cactus Preserve — an X/X green Plant creature with reach until end of turn",
					x, x, "Plant", []string{"G"}, "reach")
			},
		}},
	})
}
