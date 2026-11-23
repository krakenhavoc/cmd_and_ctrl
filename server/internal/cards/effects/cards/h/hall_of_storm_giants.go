package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hall of Storm Giants — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {U}.
//	 {5}{U}: Until end of turn, this land becomes a 7/7 blue Giant creature
//	 with ward {3}. It's still a land. (Whenever it becomes the target of a
//	 spell or ability an opponent controls, counter it unless that player
//	 pays {3}.)"
//
// The animation is afrB17Manland. Ward is a triggered ability (CR 702.21a),
// so it is the ability bundle granted with the effect (ADR 0093), as the
// quoted triggers of the other creature lands are: it exists only while the
// land is animated.
//
// No simplification.
const HallOfStormGiantsWard = "hall-of-storm-giants/ward"

func init() {
	effects.Register(effects.Spec{
		OracleID:      "087c8c0e-a91c-4e3c-8387-9312db01f343",
		Name:          "Hall of Storm Giants",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.AfrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: effects.AfrB17TapForMana("U"),
		Grants: []effects.AbilityGrant{{
			Key:       HallOfStormGiantsWard,
			Triggered: []game.TriggeredAbility{effects.Ward(effects.WardMana("{3}"), "Hall of Storm Giants — ward {3}")},
			Text:      "Ward {3}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}: Until end of turn, this land becomes a 7/7 blue Giant creature with ward {3}. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{5}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AfrB17Manland(g, item, "Hall of Storm Giants", "Giant", "U", 7, 7, nil, HallOfStormGiantsWard)
			},
		}},
	})
}
