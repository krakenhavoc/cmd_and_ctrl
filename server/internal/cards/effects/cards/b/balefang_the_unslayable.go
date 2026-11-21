package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balefang the Unslayable — Legendary Creature — Demon Dragon {3}{R}{R},
// 5/5 (Mystery Booster Commander Edition):
//
//	"Flying, trample, protection from Angels and from Archons
//	 Players can't gain life.
//	 When Balefang enters, target opponent creates a tapped Baneslayer
//	 Angel token. The token is goaded for the rest of the game."
//
// Two protections, not one (CR 702.16m). "Players can't gain life" is the
// Everlasting Torment static (CR 119.7), so the token's lifelink gains
// nothing while Balefang is out. The opponent creates the token (it is
// theirs, tapped) and Balefang's controller goads it, for the rest of the
// game: the goad does not end at the goader's next turn (CR 701.15a), only
// when the token leaves. The shared goad tail Rendmaw and Nettling Nuisance
// use goads what the creation made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "195d1f2a-0d71-475b-9dd1-0c8efe64ad8f",
		Name:            "Balefang the Unslayable",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "protection from Angels", "protection from Archons"},
		CantGainLife:    effects.PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Balefang the Unslayable — target opponent creates a tapped Baneslayer Angel token, goaded for the rest of the game", balefangGivesTheAngel),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func balefangGivesTheAngel(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	victim, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: victim,
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(effects.TokenCard("5/5 white Angel named Baneslayer Angel")),
			Count:    1,
			Entry:    game.TokenEntryOptions{Tapped: true},
		}},
	}, ctx.ItemTail(effects.Cmdr2S34GoadTokensForRestOfGame))
}
