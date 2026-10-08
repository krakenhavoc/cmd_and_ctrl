package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manifest Dread — Sorcery {1}{G}:
//
//	"Manifest dread. (Look at the top two cards of your library. Put
//	 one onto the battlefield face down as a 2/2 creature and the other
//	 into your graveyard. Turn it face up any time for its mana cost if
//	 it's a creature card.)"
//
// The keyword action itself (ADR 0082's 2026-10-07 amendment,
// CR 701.62a): the controller alone looks at the pair and names the
// one that enters. Turning it face up for its mana cost is the
// manifested row of TurnFaceUpOffer, which already exists.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df5a66d1-61f6-44d9-b464-3ee739e78dce",
		Name:         "Manifest Dread",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ManifestDread{}.Apply(ctx)
		},
	})
}
