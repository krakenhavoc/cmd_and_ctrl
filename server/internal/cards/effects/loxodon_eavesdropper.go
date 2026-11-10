package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Eavesdropper — Creature — Elephant Detective {3}{G}, 3/3
// (Murders at Karlov Manor, #4469):
//
//	"When this creature enters, investigate. (Create a Clue token. It's
//	 an artifact with "{2}, Sacrifice this token: Draw a card.")
//	 Whenever you draw your second card each turn, this creature gets
//	 +1/+1 and gains vigilance until end of turn."
//
// The enters trigger is Novice Inspector's one Clue. The draw trigger is
// Knights of Dol Amroth's condition (b41DrewYourSecondCardThisTurn); its
// body is a +1/+1 and a vigilance grant on the Eavesdropper, both until
// end of turn, and nothing if it has left by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "141a2757-8f20-4e10-9866-bb5d612ed62d",
		Name:         "Loxodon Eavesdropper",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loxodon Eavesdropper — investigate", Do(CreateToken{Template: ClueToken(), N: 1})),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Loxodon Eavesdropper — +1/+1 and vigilance until end of turn", loxodonEavesdropperPump),
		},
	})
}

// loxodonEavesdropperPump gives the Eavesdropper +1/+1 and vigilance
// until end of turn.
func loxodonEavesdropperPump(g *game.Game, item *game.StackItem) error {
	if err := thisGetsUntilEndOfTurn(1, 1, "Loxodon Eavesdropper — +1/+1 until end of turn")(g, item); err != nil {
		return err
	}
	return thisGainsKeywordUntilEndOfTurn("vigilance", "Loxodon Eavesdropper — gains vigilance until end of turn")(g, item)
}
