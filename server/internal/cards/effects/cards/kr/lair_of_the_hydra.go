package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lair of the Hydra — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {G}.
//	 {X}{G}: Until end of turn, this land becomes an X/X green Hydra
//	 creature. It's still a land. X can't be 0."
//
// The animation is afrB17Manland with the X announced on activation
// (CR 107.3, item.XValue) as both base power and toughness; "X can't be 0"
// is MinX(1), so a zero announcement is refused outright.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "126e9140-2c05-4c00-8b01-5653456c736a",
		Name:          "Lair of the Hydra",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		Replacements:  []game.ReplacementEffect{effects.AfrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: effects.AfrB17TapForMana("G"),
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{G}: Until end of turn, this land becomes an X/X green Hydra creature. It's still a land. X can't be 0.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{X}{G}"), effects.MinX(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := max(item.XValue, 0)
				return effects.AfrB17Manland(g, item, "Lair of the Hydra", "Hydra", "G", x, x, nil, "")
			},
		}},
	})
}
