package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethereal Investigator — Creature — Spirit {3}{U}, 2/3 (Murders at
// Karlov Manor Commander, #5368):
//
//	"Flying
//	 When this creature enters, investigate X times, where X is the
//	 number of opponents you have.
//	 Whenever you draw your second card each turn, create a 1/1 white
//	 Spirit creature token with flying."
//
// X is counted as the trigger resolves (CR 608.2h), the Chittering
// Witch way: opponents who have left the game are not counted. The
// second ability is Atlantean Cavalry's draw tally
// (b41DrewYourSecondCardThisTurn), so it fires on the second draw of a
// turn and no other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "24038629-7c12-45ef-8ec4-0d2bb3e2265c",
		Name:            "Ethereal Investigator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ethereal Investigator — investigate X times, where X is the number of opponents you have",
				Do(InvestigateForEach{
					Name:  "the number of opponents you have",
					Count: func(ctx *Context) int { return len(ctx.Opponents()) },
				})),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Ethereal Investigator — create a 1/1 white Spirit creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
