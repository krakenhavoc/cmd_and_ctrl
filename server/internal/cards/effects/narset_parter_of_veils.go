package effects

// Narset, Parter of Veils — Legendary Planeswalker — Narset {1}{U}{U},
// loyalty 5:
//
//	"Each opponent can't draw more than one card each turn.
//	 −2: Look at the top four cards of your library. You may reveal a
//	 noncreature, nonland card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// #3927. The static is Spec.DrawLimits (CR 121.2b, game/draw_limit.go):
// before each of an opponent's individual draws, a draw past their
// first this turn is not performed. It counts cards drawn earlier in the
// turn, before Narset arrived, and a draw a replacement turned into
// something else was not a draw (the 2019-05-03 rulings); a draw she
// stops can't be replaced (dredge, Underrealm Lich). "Draw three" draws
// one and the rest of the instruction goes on.
//
// The −2 is the Horn of the Mark sentence with a noncreature, nonland
// filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab26fbe2-e808-48b9-8d0d-3fbb6c3d554f",
		Name:            "Narset, Parter of Veils",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		DrawLimits:      EachOpponentCantDrawMoreThanOneCardEachTurn(),
		Activated: []ActivatedAbility{{
			Label: "−2: Look at the top four cards of your library. You may reveal a noncreature, nonland card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:  LoyaltyCost(-2),
			Effect: LookAtTopThenMayTakeToHand(4, And(Noncreature(), Nonland()), 1,
				"Narset, Parter of Veils — reveal a noncreature, nonland card and put it into your hand"),
		}},
	})
}
