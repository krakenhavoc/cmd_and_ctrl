package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Duras Sisters — Legendary Creature — Klingon Warrior {B/R}{B/R},
// 3/2 (Star Trek, #4973):
//
//	"When The Duras Sisters die, create a 2/1 red Klingon Warrior
//	 creature token with haste."
//
// A dies trigger (CR 700.4): it fires only when the Sisters go from the
// battlefield to a graveyard, so an exile or a bounce makes no token. The
// token is the set's printed Klingon Warrior (Cha'DIch Investigator's
// row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03973bad-1b3c-4488-a23e-232b192c95cc",
		Name:         "The Duras Sisters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("The Duras Sisters — create a 2/1 red Klingon Warrior with haste",
				Do(CreateToken{Template: TokenCard("2/1 red Klingon Warrior with haste"), N: 1})),
		},
	})
}
