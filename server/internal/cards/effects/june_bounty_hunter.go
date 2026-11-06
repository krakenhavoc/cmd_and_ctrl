package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// June, Bounty Hunter — Legendary Creature — Human Mercenary {1}{B}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"June can't be blocked as long as you've drawn two or more cards
//	 this turn.
//	 {1}, Sacrifice another creature: Create a Clue token. Activate only
//	 during your turn."
//
// The evasion is a block rule read as blocks are declared
// (CantBeBlockedWhile) off her controller's draw tally for the turn;
// cards put into a hand without the word "draw" are not in it. Once she
// is blocked, drawing more changes nothing (CR 509.1h). The ability's
// sacrifice is a cost of another creature, and making the Clue is not
// investigating.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1779869a-5e54-4ab5-86b8-d4d20e2e1ba4",
		Name:         "June, Bounty Hunter",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{CantBeBlockedWhile(OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return g.TurnTallyFor(c.Controller).CardsDrawn >= 2
		})},
		Activated: []ActivatedAbility{{
			Label:     "{1}, Sacrifice another creature: Create a Clue token. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
			Condition: DuringYourTurn(),
			Effect:    Do(CreateToken{Template: ClueToken(), N: 1}),
		}},
	})
}
