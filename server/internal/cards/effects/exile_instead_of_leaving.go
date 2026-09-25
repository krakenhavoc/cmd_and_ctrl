package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_instead_of_leaving.go — #1221: "if it would leave the
// battlefield, exile it instead of putting it anywhere else", pinned
// to ONE permanent.
//
// Two cards print the clause and both got it from a reanimation that
// is meant to be temporary: Whip of Erebos (#296) and every unearth
// card (CR 702.82a). It was written inline on the first; the second
// would have been a second copy of a clause whose failure mode is
// silent — an omitted `NewZone != ZoneExile` guard is an infinite
// loop, and a missed event kind is a creature that can be sacrificed
// back into a second reanimation. One helper, so the two cannot
// drift.
//
// FOR AS LONG AS THE OBJECT REMAINS (#1591). The printed clause has no
// duration, so the effect lasts until the permanent it names leaves
// the battlefield — which, by the clause itself, is into exile. It is a
// ScopedEffect record (ADR 0041 phase 3 tier 3b), pinned to the object
// with an indefinite duration: the engine drops it once that object is
// gone (CR 400.7), and not before. It used to live in a registry swept
// at cleanup, so a creature whose end-step exile was countered (Stifle)
// survived the turn without it, died into the graveyard, and could be
// whipped again.
//
// A bounce is redirected too: since #539 BounceToHandForEffect moves
// the card through the shared exit primitive, which runs the CR 614
// window every other battlefield exit does.

// ExileInsteadOfLeavingBattlefield registers the CR 614 replacement
// that redirects ANY battlefield exit of the permanent `cardID` to
// exile, under `controller`'s control of the replacement effect. It
// registers nothing for a card that is not on the battlefield.
//
// Caller must hold g.mu — every caller is an ability body running at
// resolution, which does.
func ExileInsteadOfLeavingBattlefield(g *game.Game, cardID, controller uuid.UUID, label string) {
	g.ExileInsteadOfLeavingBattlefieldForEffect(uuid.Nil, cardID, controller, label)
}

// ExileInsteadOfGraveyardThisTurn is "if a permanent you control would
// be put into a graveyard from the battlefield this turn, exile it
// instead" (Cosmic Intervention), with Then — a registered delayed-
// trigger body — scheduled at the beginning of the next end step for
// each permanent it exiles, carrying that card in the fired item's
// Targets. "You" is the resolving item's controller, and the set is
// read live (a replacement effect is not a characteristic, so CR 611.2c
// does not lock it): a permanent that enters after the spell resolved
// is saved too. A sacrifice is caught as well as a destruction, since
// the text says "put into a graveyard from the battlefield".
type ExileInsteadOfGraveyardThisTurn struct {
	Then  game.BodyRef
	Label string
}

func (e ExileInsteadOfGraveyardThisTurn) Apply(ctx *Context) error {
	label := e.Label
	if label == "" {
		label = "exile instead of graveyard"
	}
	ctx.Game.ExileInsteadOfGraveyardThisTurnForEffect(ctx.Source(), ctx.Controller(), e.Then, label)
	return nil
}
