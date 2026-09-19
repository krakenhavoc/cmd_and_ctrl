package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrilling Discovery — Sorcery {R}{W} (EDHREC rank 4132):
//
//	"You gain 2 life. Then you may discard two cards. If you do, draw
//	 three cards."
//
// A two-mana Cathartic Reunion with a life buffer, and the reason
// both are played: a reanimator or a Hollow One deck wants the two
// cards in the graveyard as much as it wants the three in hand, so
// the "drawback" is the payoff. The 2 life is what makes the red-white
// version playable in a deck already paying life for its lands.
//
// THE THREE CLAUSES ARE IN PRINTED ORDER AND THE ORDER IS OBSERVABLE.
// The life comes first and is unconditional — you gain it even with
// an empty hand and even if you decline. The discard is second and
// is a "you MAY": declining is legal and leaves you a two-mana
// Healing Salve, which is occasionally right when the two cards you
// would pitch are the two you need.
//
// "IF YOU DO" IS A REAL CONDITION AND IT IS ALL-OR-NOTHING. The
// printed clause is "discard TWO cards", so pitching one is not a
// smaller version of the deal — it is not on offer. The prompt is
// therefore a ceiling of two with a set-level check that refuses any
// answer of exactly one, which makes it the yes-or-no the card
// prints. A player with fewer than two cards in hand cannot take the
// deal at all and draws nothing, which is CR 701.8a applied to an
// "if you do" clause: you cannot pay half a cost and collect.
//
// The count is the prompted-discard RUN's — see
// b39MayDiscardThenDraw (#1027).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5f5a234-c751-4976-a3e1-fbd52b3255c9",
		Name:         "Thrilling Discovery",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (GainLife{Player: item.Controller, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return b39MayDiscardThenDraw(2, true,
				"Thrilling Discovery — discard two cards to draw three?",
				func(discarded int) int {
					if discarded < 2 {
						return 0
					}
					return 3
				})(ctx)
		},
	})
}
