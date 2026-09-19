package effects

// Gitaxian Probe — Sorcery {U/P} (EDHREC rank 523):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Look at target player's hand.
//	 Draw a card."
//
// A peek and a cantrip. The body is
// lookAtTargetPlayersHandThenDraw, shared with Peek: "look at" is not
// "reveal", so only the caster learns the hand, and the effect marks
// the caster as a knower of each card rather than calling
// RevealHandForEffect, which would show the hand to the whole table —
// information the printed card never gives the other two players.
//
// The Phyrexian symbol's "or 2 life" half is the ENGINE's since #787
// (a cast announcing CastSpellParams.PhyrexianLife: 1 pays 2 life
// through the cost path and owes no mana, CR 107.4c) and the BOARD's
// since #916 (the cast prompt offers "pay 1 with life"). So the free
// draw is really free of mana, which is the whole card.
func init() {
	Register(Spec{
		OracleID:     "1d67f5ff-1fce-45e5-b6a1-416c569351e2",
		Name:         "Gitaxian Probe",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    lookAtTargetPlayersHandThenDraw,
	})
}
