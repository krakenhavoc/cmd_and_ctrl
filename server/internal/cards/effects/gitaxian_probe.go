package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gitaxian Probe — Sorcery {U/P} (EDHREC rank 523):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Look at target player's hand.
//	 Draw a card."
//
// A peek and a cantrip. "Look at" is not "reveal": only the caster
// learns the hand, so the effect marks the caster as a knower of
// each card rather than calling RevealHandForEffect, which would
// show the hand to the whole table — information the printed card
// never gives the other two players.
//
// The Phyrexian symbol's "or 2 life" half is the ENGINE's since #787:
// a cast announcing CastSpellParams.PhyrexianLife: 1 pays 2 life
// through the CR 119.4 cost path and owes no mana (CR 107.4c). What
// is still missing is the BOARD's half — no button asks the question,
// so a player clicking the card from hand pays {U}. A client seam,
// not a card file's.
func init() {
	Register(Spec{
		OracleID:     "1d67f5ff-1fce-45e5-b6a1-416c569351e2",
		Name:         "Gitaxian Probe",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The board has no button for the Phyrexian symbol yet — casting it from hand pays {U}, not 2 life. The engine accepts the life payment."},
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetPlayer {
				if p := ctx.PlayerByID(item.Targets[0].ID); p != nil && p.Hand != nil {
					for i := range p.Hand.Cards {
						p.Hand.Cards[i].AddKnower(ctx.Controller())
					}
				}
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
