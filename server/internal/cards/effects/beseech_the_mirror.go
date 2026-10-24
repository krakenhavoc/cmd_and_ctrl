package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beseech the Mirror — Sorcery {1}{B}{B}{B}:
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as
//	 you cast this spell.)
//	 Search your library for a card, exile it face down, then shuffle.
//	 If this spell was bargained, you may cast the exiled card without
//	 paying its mana cost if that spell's mana value is 4 or less. Put
//	 the exiled card into your hand if it wasn't cast this way."
//
// Bargain is the shared optional cost (g2Bargain, exile_if_dies_g2.go). The
// search exiles face down, so the table sees a card back. Bargained,
// the card is offered as a free cast capped at mana value 4
// (Game.CastExiledFreeOrToHandForEffect), cast inside the spell's
// resolution (CR 608.2g, #4250): declined or over the cap, it goes into
// its owner's hand. Not bargained, it goes
// straight to the hand, a Diabolic Tutor that hides the card on the
// way.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:      "cbbe535b-12d6-4e6c-8b26-9220f06a1604",
		Name:          "Beseech the Mirror",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		Purpose:       game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			maxMV := -1
			if g2WasBargained(ctx) {
				maxMV = 4
			}
			return SearchLibrary{
				Player:   ctx.Controller(),
				Limit:    1,
				Dest:     game.ZoneExile,
				FaceDown: game.FaceDownPermitted,
				Shuffle:  true,
				Reason:   "Beseech the Mirror — search your library for a card",
				// #4253 slice 4: the mana value cap rides the tail.
				ThenTail: ctx.TailN(beseechTheMirrorCast, maxMV),
			}.Apply(ctx)
		},
	})
}

// beseechTheMirrorCast is "you may cast it without paying its mana cost if
// its mana value is Tail.N or less (any, when N is negative); put it into
// your hand if you don't". An on-disk key.
var beseechTheMirrorCast = game.RegisterTail("beseech-the-mirror/cast-or-to-hand", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.CastExiledFreeOrToHandForEffect(o.Tail.Player, o.Tail.Source, id, o.Tail.N); err != nil {
			return err
		}
	}
	return nil
})
