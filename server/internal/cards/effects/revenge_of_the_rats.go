package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revenge of the Rats — Sorcery {2}{B}{B} (Foundations, #4465):
//
//	"Create a tapped 1/1 black Rat creature token for each creature card
//	 in your graveyard.
//	 Flashback {2}{B}{B}"
//
// Dog Walker's tapped-token spec with the count read as the spell resolves
// (CR 608.2h): every creature card in the caster's graveyard, counted off
// its printed type line, since a card in a graveyard has no layers. Flashback
// is the shared alternative cost; the spell is exiled instead of going to
// the graveyard when it resolves from there (CR 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "29e3a3d9-4839-44d0-ad48-159c01645a6a",
		Name:             "Revenge of the Rats",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{B}{B}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			return CreateTokenAdvanced{Spec: Token(TokenCard("1/1 black Rat")).EntersTapped(), N: n}.Apply(ctx)
		},
	})
}
