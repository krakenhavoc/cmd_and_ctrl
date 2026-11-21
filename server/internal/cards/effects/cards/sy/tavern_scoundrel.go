package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tavern Scoundrel — Creature — Human Rogue {1}{R}, 1/3:
//
//	"Whenever you win a coin flip, create two Treasure tokens.
//	 {1}, {T}, Sacrifice another permanent: Flip a coin."
//
// The activated ability's flip prints no outcome, but the first ability
// cares whether you win it, so it is a called flip (FlipACoin, CR 705.2)
// and winning it is what makes the Treasures (WheneverYouWinACoinFlip,
// #3767). Any other coin you win makes them too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68c4f86f-ff12-44a6-8f18-f2ee9753b2cd",
		Name:         "Tavern Scoundrel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouWinACoinFlip("Tavern Scoundrel — create two Treasure tokens", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another permanent: Flip a coin.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another permanent", effects.Permanent())),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.FlipACoin{Question: "Tavern Scoundrel — call the coin flip"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
