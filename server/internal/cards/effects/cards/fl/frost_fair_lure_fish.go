package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frost Fair Lure Fish — Creature — Fish {5}{U}{R}, 7/7 (Doctor Who):
//
//	"When this creature enters, create two 1/1 blue Fish creature tokens and
//	 create two tapped Treasure tokens.
//	 Fish you control have haste and can't be blocked by Humans.
//	 Foretell {3}{U}{R} (During your turn, you may pay {2} and exile this card from
//	 your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The two Fish are the existing 1/1 blue Fish token and the Treasures arrive tapped.
// "Fish you control" is every Fish creature its controller has, this one included:
// the haste is a layer-6 grant and the evasion a block rule on the attackers (Stromkirk
// Noble's, over the Fish you control), both read through the layers. Foretell is the
// engine's special action.
//
// No simplification.
func init() {
	fish := effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Fish"))
	effects.Register(effects.Spec{
		OracleID:       "51e0a61d-69f0-475d-b40b-50d69252800d",
		Name:           "Frost Fair Lure Fish",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{3}{U}{R}")},
		Purpose:        game.Purpose{Tokens: 4, CreatureTokens: 2},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature() && target.HasSubtypeWord("Fish")
			}, "haste"),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnMatching(fish), effects.Subtype("Human"), "Humans"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Frost Fair Lure Fish — create two 1/1 blue Fish creature tokens and two tapped Treasure tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Fish"), N: 2}, effects.CreateToken{Template: effects.TappedTreasureToken(), N: 2})),
		},
	})
}
