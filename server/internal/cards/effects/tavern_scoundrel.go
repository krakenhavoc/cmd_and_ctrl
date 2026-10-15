package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "68c4f86f-ff12-44a6-8f18-f2ee9753b2cd",
		Name:         "Tavern Scoundrel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouWinACoinFlip("Tavern Scoundrel — create two Treasure tokens", Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another permanent: Flip a coin.",
			Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeAnotherN(1, "another permanent", Permanent())),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return FlipACoin{Question: "Tavern Scoundrel — call the coin flip"}.Apply(NewContext(g, item))
			},
		}},
	})
}
